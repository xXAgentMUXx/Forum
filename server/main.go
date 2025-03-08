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

	http.HandleFunc("/", auth.ServeHTML)
	http.HandleFunc("/register", auth.RegisterUser)
	http.HandleFunc("/login", auth.LoginUser)
	http.HandleFunc("/logout", auth.LogoutUser)
	http.HandleFunc("/check-session", auth.CheckSession)
	http.HandleFunc("/auth/google", auth.AuthMiddleware(auth.AuthGoogle))
	http.HandleFunc("/auth/github", auth.AuthMiddleware(auth.AuthGithub))
	http.HandleFunc("/auth/callback/google", auth.AuthMiddleware(auth.GoogleCallback))
	http.HandleFunc("/auth/callback/github", auth.AuthMiddleware(auth.GithubCallback))
	http.HandleFunc("/forum", auth.AuthMiddleware(forum.ServeForum))
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
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	fmt.Println("Server started on https://localhost:8080")

	err = http.ListenAndServeTLS(":8080", "localhost+2.pem", "localhost+2-key.pem", nil)
	if err != nil {
		log.Fatal("HTTPS Error: ", err)
	}
}


