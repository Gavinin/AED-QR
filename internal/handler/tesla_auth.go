package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	TeslaClientID    = "ownerapi"
	TeslaRedirectURI = "https://auth.tesla.com/void/callback"
	TeslaScope       = "openid email offline_access"
	TeslaAuthBaseURL = "https://auth.tesla.com/oauth2/v3/authorize"
	TeslaTokenURL    = "https://auth.tesla.com/oauth2/v3/token"
)

// GenerateTeslaAuthURL generates PKCE params and returns the auth URL
func GenerateTeslaAuthURL(c *gin.Context) {
	verifier, err := generateRandomString(86)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verifier"})
		return
	}

	challenge := generateCodeChallenge(verifier)
	state, err := generateRandomString(12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate state"})
		return
	}

	params := url.Values{}
	params.Add("client_id", TeslaClientID)
	params.Add("code_challenge", challenge)
	params.Add("code_challenge_method", "S256")
	params.Add("redirect_uri", TeslaRedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", TeslaScope)
	params.Add("state", state)

	authURL := fmt.Sprintf("%s?%s", TeslaAuthBaseURL, params.Encode())

	c.JSON(http.StatusOK, gin.H{
		"url":      authURL,
		"verifier": verifier,
		"state":    state,
	})
}

// ExchangeTeslaTokenRequest
type ExchangeTeslaTokenRequest struct {
	Code     string `json:"code" binding:"required"`
	Verifier string `json:"verifier" binding:"required"`
}

// ExchangeTeslaToken exchanges the authorization code for tokens
func ExchangeTeslaToken(c *gin.Context) {
	var req ExchangeTeslaTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Determine API URL based on code prefix (CN_ for China)
	apiURL := TeslaTokenURL
	if strings.HasPrefix(req.Code, "CN_") {
		apiURL = "https://auth.tesla.cn/oauth2/v3/token"
	}

	data := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     TeslaClientID,
		"code":          req.Code,
		"code_verifier": req.Verifier,
		"redirect_uri":  TeslaRedirectURI,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request data"})
		return
	}

	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request to Tesla"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response body"})
		return
	}

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Tesla API error: %s", string(body))})
		return
	}

	// Return raw JSON response from Tesla
	c.Data(http.StatusOK, "application/json", body)
}

// Helper functions (duplicated from token_gen/main.go but cleaner to have here)
func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// base64 url encoding without padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateCodeChallenge(verifier string) string {
	h := sha256.New()
	h.Write([]byte(verifier))
	sum := h.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(sum)
}
