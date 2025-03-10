package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth "Forum/auth"
	forum "Forum/forum"

	rate "Forum/security"

	_ "github.com/mattn/go-sqlite3"
)

func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Pas de fichier .env trouvé")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("❌ Erreur de lecture du fichier .env :", err)
	}
}

func main() {
	loadEnvFile(".env")

	auth.InitDB()
	auth.InitOAuth()

	limiter := rate.NewRateLimiter(200, 60*time.Second) 
	mux := http.NewServeMux()
	
	mux.Handle("/", limiter.Limit(http.HandlerFunc(auth.ServeHTML)))
	mux.Handle("/register", limiter.Limit(http.HandlerFunc(auth.RegisterUser)))
	mux.Handle("/login", limiter.Limit(http.HandlerFunc(auth.LoginUser)))
	mux.Handle("/logout", limiter.Limit(http.HandlerFunc(auth.LogoutUser)))
	mux.Handle("/edit_user", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(auth.EditUser))))
	mux.Handle("/check-session", limiter.Limit(http.HandlerFunc(auth.CheckSession)))
	mux.Handle("/auth/google", limiter.Limit(http.HandlerFunc((auth.AuthGoogle))))
	mux.Handle("/auth/github", limiter.Limit(http.HandlerFunc((auth.AuthGithub))))
	mux.Handle("/auth/callback/google", limiter.Limit(http.HandlerFunc((auth.GoogleCallback))))
	mux.Handle("/auth/callback/github", limiter.Limit(http.HandlerFunc((auth.GithubCallback))))
	mux.Handle("/forum", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(forum.ServeForum))))
	mux.Handle("/forum_invite", limiter.Limit(http.HandlerFunc(forum.ServeForumInvite)))
	mux.Handle("/post/create", limiter.Limit(http.HandlerFunc(forum.CreatePost)))
	mux.Handle("/posts", limiter.Limit(http.HandlerFunc(forum.GetAllPosts)))
	mux.Handle("/categories", limiter.Limit(http.HandlerFunc(forum.GetCategories)))
	mux.Handle("/comments", limiter.Limit(http.HandlerFunc(forum.GetComments)))
	mux.Handle("/like/comment", limiter.Limit(http.HandlerFunc(forum.LikeComment)))
	mux.Handle("/comment/create", limiter.Limit(http.HandlerFunc(forum.CreateComment)))
	mux.Handle("/post/delete", limiter.Limit(http.HandlerFunc(forum.DeletePost)))
	mux.Handle("/comment/delete", limiter.Limit(http.HandlerFunc(forum.DeleteComment)))
	mux.Handle("/like/post", limiter.Limit(http.HandlerFunc(forum.Like_Post)))
	mux.Handle("/likes", limiter.Limit(http.HandlerFunc(forum.GetLikesAndDislike)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	fmt.Println("✅ Serveur lancé sur https://localhost:8080") // Commande Docker :  sudo docker compose up --build

	err := http.ListenAndServeTLS(":8080", "localhost+2.pem", "localhost+2-key.pem", mux)
	if err != nil {
		log.Fatal("❌ Erreur HTTPS :", err)
	}
}
