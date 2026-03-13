package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// Constants
const (
	ClientID    = "ownerapi"
	RedirectURI = "https://auth.tesla.com/void/callback"
	Scope       = "openid email offline_access"
	AuthBaseURL = "https://auth.tesla.com/oauth2/v3/authorize"
	TokenURL    = "https://auth.tesla.com/oauth2/v3/token"
)

// TokenResponse structure
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func main() {
	// 1. Generate PKCE params
	codeVerifier, err := generateRandomString(86)
	if err != nil {
		fmt.Printf("Error generating code verifier: %v\n", err)
		return
	}

	codeChallenge := generateCodeChallenge(codeVerifier)
	state, err := generateRandomString(12)
	if err != nil {
		fmt.Printf("Error generating state: %v\n", err)
		return
	}

	// 2. Construct Authorization URL
	params := url.Values{}
	params.Add("client_id", ClientID)
	params.Add("code_challenge", codeChallenge)
	params.Add("code_challenge_method", "S256")
	params.Add("redirect_uri", RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", Scope)
	params.Add("state", state)

	authURL := fmt.Sprintf("%s?%s", AuthBaseURL, params.Encode())
	fmt.Println("Opening browser for authentication...")
	fmt.Printf("URL: %s\n\n", authURL)

	// 3. Open Browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Println("Please copy and paste the URL above into your browser.")
	}

	// 4. Prompt user to paste the redirect URL
	fmt.Println("---------------------------------------------------------")
	fmt.Println("After logging in, you will see a 'Page Not Found' error.")
	fmt.Println("Copy the full URL from your browser's address bar and paste it below:")
	fmt.Println("---------------------------------------------------------")

	var redirectURLStr string
	fmt.Print("Paste URL here: ")
	fmt.Scanln(&redirectURLStr)

	// 5. Extract Code
	u, err := url.Parse(redirectURLStr)
	if err != nil {
		fmt.Printf("Invalid URL: %v\n", err)
		return
	}

	code := u.Query().Get("code")
	returnedState := u.Query().Get("state")

	if code == "" {
		fmt.Println("Error: No code found in URL")
		return
	}

	if returnedState != state {
		fmt.Println("Warning: State mismatch! Potential CSRF attack.")
		// Proceeding anyway for manual tool but this is a security risk in production
	}

	fmt.Printf("Authorization Code: %s\n", code)

	// 6. Exchange Code for Token
	fmt.Println("Exchanging code for token...")
	tokens, err := exchangeCodeForToken(code, codeVerifier)
	if err != nil {
		fmt.Printf("Error exchanging token: %v\n", err)
		return
	}

	// 7. Output Result
	fmt.Println("\n---------------------------------------------------------")
	fmt.Println("Authentication Successful!")
	fmt.Println("---------------------------------------------------------")
	fmt.Printf("Access Token: %s\n", tokens.AccessToken)
	fmt.Printf("Refresh Token: %s\n", tokens.RefreshToken)
	fmt.Printf("Expires In: %d seconds\n", tokens.ExpiresIn)

	// Create JSON output for config
	configOutput := map[string]string{
		"Access Token":  tokens.AccessToken,
		"Refresh Token": tokens.RefreshToken,
	}
	jsonBytes, _ := json.MarshalIndent(configOutput, "", "  ")
	fmt.Println("\nJSON for Config:")
	fmt.Println(string(jsonBytes))
}

// Helper Functions

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

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	cmdExec := exec.Command(cmd, args...)
	return cmdExec.Start()
}

func exchangeCodeForToken(code, verifier string) (*TokenResponse, error) {
	// API endpoint
	apiURL := "https://auth.tesla.com/oauth2/v3/token"
	if strings.HasPrefix(code, "CN_") {
		apiURL = "https://auth.tesla.cn/oauth2/v3/token"
	}

	data := map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     ClientID,
		"code":          code,
		"code_verifier": verifier,
		"redirect_uri":  RedirectURI,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned error: %s - %s", resp.Status, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}
