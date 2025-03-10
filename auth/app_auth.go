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

var GoogleOauthConfig *oauth2.Config
var GithubOauthConfig *oauth2.Config

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

func AuthGoogle(w http.ResponseWriter, r *http.Request) {
	url := GoogleOauthConfig.AuthCodeURL("state")
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func AuthGithub(w http.ResponseWriter, r *http.Request) {
	url := GithubOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

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
	setSessionCookie(w, email)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

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
	if !ok || email == "" {
		fmt.Println("❌ [DEBUG] Email manquant dans la réponse GitHub. Tentative de récupération des emails...")
		
		resp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			http.Error(w, "Failed to get user emails", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		var emails []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&emails)

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
	setSessionCookie(w, email)
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}
func setSessionCookie(w http.ResponseWriter, email string) {
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{
		Name:     "session",
		Value:    email,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true, 
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}