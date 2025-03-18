package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

// Auth configurations for Google and GitHub
var GoogleOauthConfig *oauth2.Config
var GithubOauthConfig *oauth2.Config

// initializes the OAuth configurations
func InitOAuth() {
	GoogleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "https://localhost:8080/auth/callback/google",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
	GithubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "https://localhost:8080/auth/callback/github",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}
}

// AuthGoogle initiates the Google OAuth
func AuthGoogle(w http.ResponseWriter, r *http.Request) {
// Generate the Google login URL and redirect the user
	url := GoogleOauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// AuthGithub initiates the Github OAuth
func AuthGithub(w http.ResponseWriter, r *http.Request) {
	// Generate the Google login URL and redirect the user
	url := GithubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the callback
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Retrieve the authorization code 
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	// Exchange the authorization code for an access token
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	// Use the access token to fetch user information 
	client := GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the response and retrieve the user's email
	var userInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&userInfo)

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		http.Error(w, "No email found in Google response", http.StatusUnauthorized)
		return
	}
	// Set a session cookie with the user's email and redirect to the forum
	setSessionCookie(w, email)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// GithubCallback handles the callback 
func GithubCallback(w http.ResponseWriter, r *http.Request) {
	// Retrieve the authorization code
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	// Exchange the authorization code for an access token
	token, err := GithubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	// Use the access token to fetch user
	client := GithubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()


	// Parse the response and retrieve the user's email
	var userInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&userInfo)

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		fmt.Println("❌ [DEBUG] Email manquant dans la réponse GitHub. Tentative de récupération des emails...")
		
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get user emails", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Parse the list of emails and find the primary email
		var emails []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&emails)

		// Look for the primary email and use it
		for _, e := range emails {
			if primary, ok := e["primary"].(bool); ok && primary {
				email, _ = e["email"].(string)
				break
			}
		}
	}
	if email == "" {
		http.Error(w, "No email found in GitHub response", http.StatusUnauthorized)
		return
	}
	// Set a session cookie with the user's email and redirect to the forum
	setSessionCookie(w, email)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}
// sets a session cookie
func setSessionCookie(w http.ResponseWriter, email string) {
	// Set cookie expiration time
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session",
		Value:    email,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true, 
		Path:     "/",
	}
	// Set the cookie in the HTTP response
	http.SetCookie(w, &cookie)
}