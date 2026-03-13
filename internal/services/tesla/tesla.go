package tesla

import (
	"AED-QR/internal/config"
	"AED-QR/internal/initial"
	"AED-QR/internal/log"
	"AED-QR/internal/model"
	"AED-QR/internal/services"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Tesla struct {
	Params  *Params
	Vehicle *model.Vehicle // Reference to update DB/Cache
}

type Params struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	API          string `json:"api"` // auth.tesla.cn or auth.tesla.com
	VIN          string `json:"vin"`
	ID           string `json:"id"`
}

// Product represents a vehicle returned by Tesla API
type Product struct {
	ID          uint64 `json:"id"`
	VehicleID   uint64 `json:"vehicle_id"`
	VIN         string `json:"vin"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
}

type ProductsResponse struct {
	Response []Product `json:"response"`
	Count    int       `json:"count"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

// VehicleDataResponse structure for vehicle_data endpoint
type VehicleDataResponse struct {
	Response struct {
		State        string `json:"state"` // online, offline, asleep
		VehicleState struct {
			Locked             bool `json:"locked"`
			CenterDisplayState int  `json:"center_display_state"` // 0=Off, 2=On, etc.
		} `json:"vehicle_state"`
	} `json:"response"`
}

// NewTesla creates a new Tesla adapter
func NewTesla(vehicle *model.Vehicle) (*Tesla, error) {
	var p Params
	if err := json.Unmarshal([]byte(vehicle.Data), &p); err != nil {
		return nil, fmt.Errorf("failed to parse tesla params: %v", err)
	}
	return &Tesla{
		Params:  &p,
		Vehicle: vehicle,
	}, nil
}

// wakeUp wakes up the vehicle
func (t *Tesla) wakeUp() error {
	_, err := t.makeRequest("POST", fmt.Sprintf("/api/1/vehicles/%s/wake_up", t.Params.ID), nil)
	return err
}

func (t *Tesla) IsRunning() bool {
	// 1. Get Vehicle Data
	// Since vehicle_data returns all state, we can check everything.
	// However, if the vehicle is asleep or offline, this might fail or return state="offline"

	// We'll try up to 3 times if we need to wake up
	maxRetries := 30 // Wait up to 30-60 seconds if waking up
	// Actually, let's keep it simple: Try once, if offline/asleep -> WakeUp -> Wait -> Retry

	for i := 0; i < maxRetries; i++ {
		data, err := t.makeRequest("GET", fmt.Sprintf("/api/1/vehicles/%s/vehicle_data", t.Params.ID), nil)
		if err != nil {
			// If error is 408 Request Timeout or similar, it might be sleeping.
			// But makeRequest returns error string.
			log.Errorf("Error getting vehicle data: %v", err)
			// Proceed to try wake up just in case
		}

		var vData VehicleDataResponse
		// Even if err != nil, data might be empty.
		if len(data) > 0 {
			if err := json.Unmarshal(data, &vData); err == nil {
				// Check state
				state := vData.Response.State
				if state == "online" {
					// Check running conditions
					// If locked or owner left (center_display_state == 0 implies off/away?)
					// User said: "If locked or owner left then isrunning is false"
					// We interpret "owner left" as center_display_state == 0 (screen off) or similar.
					// Actually, "Locked" is the strongest indicator for "Safe to open AED" if we consider "IsRunning" as "Occupied/Driving".
					// Wait, the user logic is:
					// IsRunning = true (means busy/driving, forbid open)
					// IsRunning = false (means parked/safe, allow open)

					// User: "If locked or owner left then isrunning is false"
					// So: Locked=true OR OwnerLeft=true => IsRunning=false
					// Otherwise => IsRunning=true

					locked := vData.Response.VehicleState.Locked
					displayState := vData.Response.VehicleState.CenterDisplayState

					// We assume displayState == 0 means off/owner left.
					// If Locked is true, IsRunning is false.
					// If DisplayState is 0, IsRunning is false.
					if locked || displayState == 0 {
						return false
					}

					// If Unlocked AND Display On => It might be running/driving/occupied
					return true
				}
			}
		}

		// If we are here, vehicle is not online or request failed.
		// Trigger Wake Up
		if i == 0 {
			log.Info("Vehicle not online, sending wake_up command...")
			if err := t.wakeUp(); err != nil {
				log.Errorf("Failed to wake up vehicle: %v", err)
			}
		}

		// Wait before retry
		time.Sleep(2 * time.Second)
	}

	// If timeout/offline, we assume it's NOT running (safe default?)
	// User said "If vehicle offline, call wake_up method".
	// If we still can't get data, we probably shouldn't block the AED.
	// So return false (not running).
	return false
}

// getOwnerAPIUrl returns the base URL for Owner API based on auth region
func (t *Tesla) getOwnerAPIUrl() string {
	if strings.Contains(t.Params.API, "cn") {
		return "https://owner-api.vn.cloud.tesla.cn"
	}
	return "https://owner-api.vn.teslamotors.com"
}

// refreshToken refreshes the access token
func (t *Tesla) refreshToken() error {
	url := "https://auth.tesla.com/oauth2/v3/token"
	if strings.Contains(t.Params.API, "cn") {
		url = "https://auth.tesla.cn/oauth2/v3/token"
	}

	payload := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     "ownerapi",
		"refresh_token": t.Params.RefreshToken,
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Mark auth error
		t.Vehicle.AuthError = true
		t.updateVehicleData()
		return fmt.Errorf("refresh token failed: %s", resp.Status)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	// Update params
	t.Params.AccessToken = tokenResp.AccessToken
	t.Params.RefreshToken = tokenResp.RefreshToken

	// Reset auth error
	t.Vehicle.AuthError = false

	return t.updateVehicleData()
}

// updateVehicleData saves the updated tokens to DB and Cache
func (t *Tesla) updateVehicleData() error {
	newData, err := json.Marshal(t.Params)
	if err != nil {
		return err
	}
	t.Vehicle.Data = string(newData)

	// Update DB
	if err := initial.DB.Save(t.Vehicle).Error; err != nil {
		log.Errorf("Failed to update vehicle in DB: %v", err)
		return err
	}

	// Update Cache
	services.UpdateVehicleCache(*t.Vehicle)
	return nil
}

// makeRequest makes an API request, handling token refresh
func (t *Tesla) makeRequest(method, path string, body interface{}) ([]byte, error) {
	apiURL := t.getOwnerAPIUrl() + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	doReq := func(token string) (*http.Response, error) {
		req, err := http.NewRequest(method, apiURL, bodyReader)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "AED-QR/1.0")

		client := &http.Client{Timeout: 30 * time.Second}
		return client.Do(req)
	}

	resp, err := doReq(t.Params.AccessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		log.Info("Token expired, refreshing...")
		if err := t.refreshToken(); err != nil {
			return nil, err
		}
		// Retry with new token
		// Need to recreate body reader if it was consumed
		if body != nil {
			jsonBody, _ := json.Marshal(body)
			bodyReader = bytes.NewBuffer(jsonBody)
		}
		resp, err = doReq(t.Params.AccessToken)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

func (t *Tesla) Open() error {
	if t.Vehicle.AuthError {
		return fmt.Errorf("auth error: please contact owner %s", config.AppConfig.Admin.OwnerPhone)
	}

	// Determine command based on location
	var err error
	switch t.Vehicle.Location {
	case "Front Trunk":
		// actuate_trunk with which_trunk=front
		_, err = t.makeRequest("POST", fmt.Sprintf("/api/1/vehicles/%s/command/actuate_trunk", t.Params.ID), map[string]string{"which_trunk": "front"})
	case "Rear Trunk":
		// actuate_trunk with which_trunk=rear
		_, err = t.makeRequest("POST", fmt.Sprintf("/api/1/vehicles/%s/command/actuate_trunk", t.Params.ID), map[string]string{"which_trunk": "rear"})
	default:
		// Unlock doors
		_, err = t.makeRequest("POST", fmt.Sprintf("/api/1/vehicles/%s/command/door_unlock", t.Params.ID), nil)
	}

	if err != nil {
		// If auth error occurred during makeRequest (refresh failed), we need to check
		if t.Vehicle.AuthError {
			return fmt.Errorf("auth error: please contact owner %s", config.AppConfig.Admin.OwnerPhone)
		}
		return err
	}

	return nil
}

func (t *Tesla) Lock() error {
	if t.Vehicle.AuthError {
		return fmt.Errorf("auth error: please contact owner %s", config.AppConfig.Admin.OwnerPhone)
	}

	_, err := t.makeRequest("POST", fmt.Sprintf("/api/1/vehicles/%s/command/door_lock", t.Params.ID), nil)
	if err != nil {
		if t.Vehicle.AuthError {
			return fmt.Errorf("auth error: please contact owner %s", config.AppConfig.Admin.OwnerPhone)
		}
		return err
	}
	return nil
}

// ListVehicles fetches vehicles using provided tokens and region
// This is a static method helper that doesn't rely on an existing Vehicle instance in DB
func ListVehicles(accessToken, refreshToken, apiRegion string) ([]Product, string, string, error) {
	// Construct a temporary Tesla object just for using makeRequest logic (or duplicate logic)
	// Duplicating logic is safer to avoid creating a fake model.Vehicle

	ownerAPI := "https://owner-api.vn.teslamotors.com"
	if strings.Contains(apiRegion, "cn") {
		ownerAPI = "https://owner-api.vn.cloud.tesla.cn"
	}

	doReq := func(token string) (*http.Response, error) {
		req, err := http.NewRequest("GET", ownerAPI+"/api/1/products", nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("User-Agent", "AED-QR/1.0")

		client := &http.Client{Timeout: 30 * time.Second}
		return client.Do(req)
	}

	resp, err := doReq(accessToken)
	if err != nil {
		return nil, accessToken, refreshToken, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		// Try refresh
		authURL := "https://auth.tesla.com/oauth2/v3/token"
		if strings.Contains(apiRegion, "cn") {
			authURL = "https://auth.tesla.cn/oauth2/v3/token"
		}

		payload := map[string]string{
			"grant_type":    "refresh_token",
			"client_id":     "ownerapi",
			"refresh_token": refreshToken,
		}
		jsonPayload, _ := json.Marshal(payload)

		refreshReq, _ := http.NewRequest("POST", authURL, bytes.NewBuffer(jsonPayload))
		refreshReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		refreshResp, err := client.Do(refreshReq)
		if err != nil {
			return nil, accessToken, refreshToken, err
		}
		defer refreshResp.Body.Close()

		if refreshResp.StatusCode != http.StatusOK {
			return nil, accessToken, refreshToken, fmt.Errorf("refresh failed")
		}

		var tokenResp TokenResponse
		if err := json.NewDecoder(refreshResp.Body).Decode(&tokenResp); err != nil {
			return nil, accessToken, refreshToken, err
		}

		accessToken = tokenResp.AccessToken
		refreshToken = tokenResp.RefreshToken

		// Retry list
		resp, err = doReq(accessToken)
		if err != nil {
			return nil, accessToken, refreshToken, err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, accessToken, refreshToken, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var prodResp ProductsResponse
	if err := json.NewDecoder(resp.Body).Decode(&prodResp); err != nil {
		return nil, accessToken, refreshToken, err
	}

	// Filter for cars (where vin is present)
	var cars []Product
	for _, p := range prodResp.Response {
		if p.VIN != "" {
			cars = append(cars, p)
		}
	}

	return cars, accessToken, refreshToken, nil
}
