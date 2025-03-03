package main

import (
	"fmt"
	"log"
	"net/http"

	auth "Forum/auth"
	forum "Forum/forum"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
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
	http.HandleFunc("/auth/callback/google", auth.GoogleCallback)
	http.HandleFunc("/auth/callback/github", auth.GithubCallback)
	http.HandleFunc("/forum", forum.ServeForum)
	http.HandleFunc("/forum_invite", forum.ServeForumInvite)
	http.HandleFunc("/post/create", forum.CreatePost)
	http.HandleFunc("/posts", forum.GetAllPosts)
	http.HandleFunc("/categories", forum.GetCategories)
	http.HandleFunc("/comments", forum.GetComments)
	http.HandleFunc("/like/comment", forum.LikeComment)
	http.HandleFunc("/comment/create", forum.CreateComment)
	http.HandleFunc("/post/delete", forum.DeletePost)
	http.HandleFunc("/comment/delete", forum.DeleteComment)
	http.HandleFunc("/like/post", forum.Like_Post)
	http.HandleFunc("/likes", forum.GetLikesAndDislike)
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
