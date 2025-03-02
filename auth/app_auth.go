package auth

import (
	"net/http"
	"os"

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
		RedirectURL:  "http://localhost:8080/auth/callback/google",
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
	GithubOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8080/auth/callback/github",
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
