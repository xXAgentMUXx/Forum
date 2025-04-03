package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
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
	url := GoogleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// AuthGithub initiates the Github OAuth
func AuthGithub(w http.ResponseWriter, r *http.Request) {
    url := GithubOauthConfig.AuthCodeURL("state") + "&prompt=consent"
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// GoogleCallback handles the callback
func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	client := GoogleOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&userInfo)
	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		http.Error(w, "No email found in Google response", http.StatusUnauthorized)
		return
	}

	// Check if user exist in the database
	var userID string
	err = DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// if not then creates it
		userID = uuid.New().String()
		_, err = DB.Exec("INSERT INTO users (id, email, username) VALUES (?, ?, ?)", userID, email, email)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create a session with the ID
	createUserSession(w, userID)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// GithubCallback handles the callback
func GithubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}
	token, err := GithubOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}
	client := GithubOauthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&userInfo)
	email, ok := userInfo["email"].(string)

	// If email is not present, then retrieves
	if !ok || email == "" {
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get user emails", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var emails []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&emails)

		// Search the principal email
		for _, e := range emails {
			if primary, ok := e["primary"].(bool); ok && primary {
				email, _ = e["email"].(string)
				break
			}
		}
	}
	// if email not found then error
	if email == "" {
		http.Error(w, "No email found in GitHub response", http.StatusUnauthorized)
		return
	}

	// Check if user exist
	var userID string
	err = DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err == sql.ErrNoRows {
		// if not, then creates it
		userID = uuid.New().String()
		_, err = DB.Exec("INSERT INTO users (id, email, username) VALUES (?, ?, ?)", userID, email, email)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Create session with user ID
	createUserSession(w, userID)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

// Create session based on the ID
func createUserSession(w http.ResponseWriter, userID string) {
	sessionID := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)

	_, err := DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiration)
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
}