package main

import (
	"fmt"
	"log"
	"net/http"
	
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	auth "Forum/auth"
)

func main() {
	err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }
	auth.InitDB()
	auth.InitOAuth()
	fmt.Println("GitHub Client ID:", auth.GithubOauthConfig.ClientID)
	http.HandleFunc("/", auth.ServeHTML)
	http.HandleFunc("/register", auth.RegisterUser)
	http.HandleFunc("/login", auth.LoginUser)
	http.HandleFunc("/auth/google", auth.AuthGoogle)
	http.HandleFunc("/auth/github", auth.AuthGithub)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
