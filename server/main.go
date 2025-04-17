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
	_ "github.com/xeodou/go-sqlcipher"

	
)

// Function to load the env file
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Pas de fichier .env trouvé")
		return
	}
	defer file.Close()

	// Use a scanner to read the file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		//Skip lines that are comments to have no errors
		if strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		os.Setenv(key, value)
	}
	// Check for errors
	if err := scanner.Err(); err != nil {
		fmt.Println("❌ Erreur de lecture du fichier .env :", err)
	}
}

// Function that start the server
func main() {
	loadEnvFile(".env")

	// Initialize the authentication and database
	auth.InitDB()
	auth.InitOAuth()

	// Create a rate limiter
	limiter := rate.NewRateLimiter(200, 60*time.Second) 
	// Create a new HTTP multiplexer
	mux := http.NewServeMux()
	
	// Define routes and associate them with aithmidlleware and with rate limiting
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
	mux.Handle("/admin", limiter.Limit(auth.AuthMiddleware(auth.RoleMiddleware("admin", forum.ServeAdmin))))
	mux.Handle("/request-moderator", limiter.Limit(http.HandlerFunc(forum.RequestModerator)))
	mux.Handle("/moderator-requests", limiter.Limit(http.HandlerFunc(forum.GetModeratorRequests)))
	mux.Handle("/approve-moderator", limiter.Limit(http.HandlerFunc(forum.ApproveModerator)))
	mux.Handle("/reject-moderator", limiter.Limit(http.HandlerFunc(forum.RejectModerator)))
	mux.Handle("/update-role", limiter.Limit(http.HandlerFunc(forum.UpdateUserRole)))
	mux.Handle("/remove-moderator-role", limiter.Limit(http.HandlerFunc(forum.RemoveModeratorRole)))
	mux.Handle("/get-moderators", limiter.Limit(http.HandlerFunc(forum.GetModerators)))
	mux.Handle("/post/delete_admin", limiter.Limit(http.HandlerFunc(forum.DeletePostByAdmin)))
	mux.Handle("/comments/delete_admin", limiter.Limit(http.HandlerFunc(forum.DeleteCommentAdmin)))
	mux.Handle("/report/post", limiter.Limit(http.HandlerFunc(forum.ReportPost)))
	mux.Handle("/report", limiter.Limit(http.HandlerFunc(forum.GetReports)))
	mux.Handle("/report/resolve", limiter.Limit(http.HandlerFunc(forum.ResolveReport)))
	mux.Handle("/report/reject", limiter.Limit(http.HandlerFunc(forum.RejectReport)))
	mux.Handle("/moderator", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(auth.RoleMiddleware("moderator",forum.ServeModerator)))))
	mux.Handle("/forum", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(forum.ServeForum))))
	mux.Handle("/notifications", limiter.Limit(http.HandlerFunc(forum.GetNotifications)))
	mux.Handle("/notifications/mark-seen", limiter.Limit(http.HandlerFunc(forum.MarkNotificationsAsSeen)))
	mux.Handle("/notifications/delete", limiter.Limit(http.HandlerFunc(forum.DeleteNotification)))
	mux.Handle("/activity", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(auth.ServeActivity))))
	mux.Handle("/user/activity", limiter.Limit(http.HandlerFunc(auth.AuthMiddleware(auth.GetUserActivity))))
	mux.Handle("/comments/new", limiter.Limit(http.HandlerFunc(forum.GetNewComments)))
	mux.Handle("/forum_invite", limiter.Limit(http.HandlerFunc(forum.ServeForumInvite)))
	mux.Handle("/post/create", limiter.Limit(http.HandlerFunc(forum.CreatePost)))
	mux.Handle("/posts", limiter.Limit(http.HandlerFunc(forum.GetAllPosts)))
	mux.Handle("/categories", limiter.Limit(http.HandlerFunc(forum.GetCategories)))
	mux.Handle("/categories/create", limiter.Limit(http.HandlerFunc(forum.CreateCategory)))
	mux.Handle("/categories/delete", limiter.Limit(http.HandlerFunc(forum.DeleteCategory)))
	mux.Handle("/comments", limiter.Limit(http.HandlerFunc(forum.GetComments)))
	mux.Handle("/like/comment", limiter.Limit(http.HandlerFunc(forum.LikeComment)))
	mux.Handle("/comment/create", limiter.Limit(http.HandlerFunc(forum.CreateComment)))
	mux.Handle("/post/delete", limiter.Limit(http.HandlerFunc(forum.DeletePost)))
	mux.Handle("/comment/delete", limiter.Limit(http.HandlerFunc(forum.DeleteComment)))
	mux.Handle("/like/post", limiter.Limit(http.HandlerFunc(forum.Like_Post)))
	mux.Handle("/likes", limiter.Limit(http.HandlerFunc(forum.GetLikesAndDislike)))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	// Print a message indicating the server is running (debug)
	fmt.Println("✅ Serveur lancé sur https://localhost:8080") // Commande Docker :  sudo docker compose up --build

	// Start the server on port 8080 with the provided certificates for HTTPS
	err := http.ListenAndServeTLS(":8080", "localhost+2.pem", "localhost+2-key.pem", mux)
	if err != nil {
		log.Fatal("❌ Erreur HTTPS :", err)
	}
}
