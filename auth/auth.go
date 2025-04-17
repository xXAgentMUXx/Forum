package auth

import (
	"Forum/security"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Database object
var DB *sql.DB

// InitDB initializes the SQLCipher-encrypted database
func InitDB() {
	if _, err := os.Stat("forum_encrypted.db"); os.IsNotExist(err) {
		panic(err)
	}
	dsn := "forum_encrypted.db?_key=Mathys2006"
	var err error
	DB, err = sql.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}
	_, err = DB.Exec("PRAGMA key = 'Mathys2006';")
	if err != nil {
		panic(err)
	}
	if err = DB.Ping(); err != nil {
		panic(err)
	}
	DB.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id TEXT,
		expires_at DATETIME
	)`) 
    DB.Exec(`CREATE TABLE IF NOT EXISTS rate_limit (
        user_id TEXT PRIMARY KEY,
        last_request_time DATETIME,
        request_count INTEGER
    )`)
    DB.Exec(`CREATE TABLE IF NOT EXISTS post_images (
		post_id TEXT NOT NULL,
		image_path TEXT NOT NULL,
		FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
	)`) 
    DB.Exec(`CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE,
        username TEXT,
        password TEXT
    )`)
	DB.Exec(`CREATE TABLE IF NOT EXISTS notifications (
	id TEXT PRIMARY KEY,
	user_id TEXT NOT NULL,
	post_id TEXT,
	action TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	seen BOOLEAN DEFAULT FALSE,
	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    )`)
    DB.Exec(`
    CREATE TABLE IF NOT EXISTS reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id TEXT NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
    );
    `)
    DB.Exec(`
    CREATE TABLE IF NOT EXISTS promotion_requests (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id     TEXT NOT NULL,
    status      TEXT CHECK(status IN ('pending', 'approved', 'rejected')) DEFAULT 'pending',
    requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_by TEXT NULL,
    approved_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (approved_by) REFERENCES users(id)
    );
    `)
    DB.Exec(`
    CREATE TABLE users (
    id          TEXT PRIMARY KEY,
    email       TEXT UNIQUE NOT NULL,
    username    TEXT UNIQUE NOT NULL,
    password    TEXT NULL,
    role        TEXT CHECK(role IN ('guest', 'user', 'moderator', 'admin')) DEFAULT 'user',
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    `)
}


// Creat the template with the html file and URL
func ServeHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/"+r.URL.Path)
}

// Function to handles user registration
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/html/register.html")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Retrieve form data users
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	var exists string
	// Check if the email already exists in the database
	err := DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&exists)
	if err == nil {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	}
	// Hash the user's password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error encrypting password", http.StatusInternalServerError)
		return
	}
	// Generate a new UUID
	userID := uuid.New().String()
	// Insert the new user into the database
	_, err = DB.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error registering user (pseudo already used)", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// loginLimiter for rate limit in login system
var loginLimiter = security.NewLoginLimiter()

// Function to handles user login
func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/html/login.html")
		return
	}
	// Check if the IP is blocked due to too many failed login attempts
	ip := r.RemoteAddr
	if blocked, remaining := loginLimiter.CheckLock(ip); blocked {
		http.Error(w, fmt.Sprintf("Trop de tentatives. Réessayez dans %v secondes.", int(remaining.Seconds())), http.StatusTooManyRequests)
		return
	}
	// Retrieve form data for login
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Query the database to get the user’s information
	var userID, storedPassword string
	err := DB.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
	// If there’s an error, simulate a delay to prevent brute force attacks
	if err != nil {
		time.Sleep(4 * time.Second)
		timeout := loginLimiter.FailedAttempt(ip)
		if timeout > 0 {
			http.Error(w, fmt.Sprintf("Trop de tentatives. Réessayez dans %v secondes.", int(timeout.Seconds())), http.StatusTooManyRequests)
		} else {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		}
		return
	}
	// Compare the provided password with the stored password
	if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
		time.Sleep(4 * time.Second)
		timeout := loginLimiter.FailedAttempt(ip)
		// If the password is incorrect, simulate a delay and handle the login rate limit
		if timeout > 0 {
			http.Error(w, fmt.Sprintf("Trop de tentatives. Réessayez dans %v secondes.", int(timeout.Seconds())), http.StatusTooManyRequests)
		} else {
			http.Error(w, "Invalid Password", http.StatusUnauthorized)
		}
		return
	}
	// Reset the login attempt counter for the user
	loginLimiter.Reset(ip)
	// Remove any rate limit data associated with the user
	DB.Exec("DELETE FROM rate_limit WHERE user_id = ?", userID)

	// Create a new session for the user
	sessionID := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)

	// Insert the session data into the database
	_, err = DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiration)
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}
	// Set the session token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

//Function to retrieves the user ID from the session cookie
func GetUserFromSession(r *http.Request) (string, error) {
	// Look for the session token cookie
    cookie, err := r.Cookie("session_token")
    if err == nil {
        var userID string
		 // Get the user ID from the session in the database
        err = DB.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)
        if err == nil {
            return userID, nil
        }
    }
    //Check for OAuth session cookie
    oauthCookie, err := r.Cookie("session")
    if err == nil && oauthCookie.Value != "" {
        return oauthCookie.Value, nil 
    }

    return "", fmt.Errorf("No valid session found")
}

//Function to retrieves the role from the session cookie
func GetUserFromSessionRole(r *http.Request) (string, string, error) {
    cookie, err := r.Cookie("session_token")
    if err == nil {
        var userID, role string
        // Request modifiy to retries user for the role
        err = DB.QueryRow(`
            SELECT users.id, users.role 
            FROM users 
            JOIN sessions ON users.id = sessions.user_id 
            WHERE sessions.id = ?`, cookie.Value).Scan(&userID, &role)

        if err == nil {
            return userID, role, nil
        }
    }
    // vy default we use user for the role
    oauthCookie, err := r.Cookie("session")
    if err == nil && oauthCookie.Value != "" {
        return oauthCookie.Value, "user", nil 
    }
    return "", "", fmt.Errorf("No valid session found")
}

// Function to deletes expired sessions from the database
func CleanupExpiredSessions() {
	DB.Exec("DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP")
}

// Function to handles user logout and clears session data
func LogoutUser(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err == nil {
        DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
    }
	 // Clear the session token and OAuth cookies
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    "",
        Expires:  time.Now().Add(-1 * time.Hour),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    http.SetCookie(w, &http.Cookie{
        Name:     "session",
        Value:    "",
        Expires:  time.Now().Add(-1 * time.Hour),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })

    fmt.Println("✅ Déconnexion réussie. Redirection vers l'accueil")
    http.Redirect(w, r, "/", http.StatusFound)
}

// Function to checks if the user is authenticated by verifying their session
func CheckSession(w http.ResponseWriter, r *http.Request) {
    var userID, role string
    var err error

    // Check the session with user ID
    userID, err = GetUserFromSession(r)
    if err == nil {
        // If session exist then check the role
        _, role, err = GetUserFromSessionRole(r)
        if err != nil {
            role = "user" // if we can't retrievesthe role, we put user in default
        }
        // Sent the reponse with user ID and role
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "userID": userID,
            "role":   role,
        })
        return
    }
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
// This function is a middleware that checks if the user is authenticated before allowing access
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("🔍 AuthMiddleware: Vérification en cours...")
        userID, err := GetUserFromSession(r)
        if err == nil {
            fmt.Println("✅ AuthMiddleware: Accès accordé à l'utilisateur", userID)
            next(w, r)
            return
        }
        fmt.Println("❌ AuthMiddleware: Aucun utilisateur authentifié, redirection vers l'accueil")
        http.Redirect(w, r, "/", http.StatusFound)
    }
}

// Structure to store user activity
type Activity struct {
	Posts      []Post      `json:"posts"`
	Likes      []LikeInfo  `json:"likes"`
	Comments   []CommentInfo `json:"comments"`
    CommentLikes     []CommentLikeInfo `json:"comment_likes"`
}

type Post struct {
	ID        string   `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type LikeInfo struct {
	PostID    string  `json:"post_id"`
	Title     string `json:"title"`
	Type      string `json:"type"` 
}

type CommentInfo struct {
	PostID    string   `json:"post_id"`
	Title     string `json:"title"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}
type CommentLikeInfo struct {
    CommentID string `json:"comment_id"`
    Comment   string `json:"comment"`
    PostTitle string `json:"post_title"`
    Type      string `json:"type"` 
}

// function to display the templates
func ServeActivity(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/activity.html")
}


// Function to retrieves the activity of the user
func GetUserActivity(w http.ResponseWriter, r *http.Request) {
	 // Retrieve userID
    userID, err := GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
        return
    }
    var activity Activity

	// Check if the database is initialized
	if DB == nil {
        http.Error(w, "Erreur interne : base de données non initialisée", http.StatusInternalServerError)
        return
    }
    // Fetch posts created by the user
    rows, err := DB.Query("SELECT id, title, content, created_at FROM posts WHERE user_id = ?", userID)
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des posts", http.StatusInternalServerError)
        return
    }

    defer rows.Close()

	// Loop through the fetched posts
    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
			http.Error(w, "Erreur lors de la lecture des posts", http.StatusInternalServerError)
			return
		}
        activity.Posts = append(activity.Posts, post)
    }
    // Fetch likes/dislikes 
    rows, err = DB.Query(`
        SELECT p.id, p.title, l.type
        FROM likes l
        JOIN posts p ON l.post_id = p.id
        WHERE l.user_id = ?`, userID)
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des likes", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

	// Loop through the fetched likes
    for rows.Next() {
        var like LikeInfo
        if err := rows.Scan(&like.PostID, &like.Title, &like.Type); err != nil {
            http.Error(w, "Erreur lors de la lecture des likes", http.StatusInternalServerError)
            return
        }
        activity.Likes = append(activity.Likes, like)
    }
   // Fetch comments made by the user
    rows, err = DB.Query(`
        SELECT c.post_id, p.title, c.content, c.created_at
        FROM comments c
        JOIN posts p ON c.post_id = p.id
        WHERE c.user_id = ?`, userID)
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

	// Loop through the fetched comments
    for rows.Next() {
        var comment CommentInfo
        if err := rows.Scan(&comment.PostID, &comment.Title, &comment.Comment, &comment.CreatedAt); err != nil {
            http.Error(w, "Erreur lors de la lecture des commentaires", http.StatusInternalServerError)
            return
        }
        activity.Comments = append(activity.Comments, comment)
    }
    // Fetch likes/dislike comments made by the user
    rows, err = DB.Query(`
    SELECT c.id, c.content, p.title, l.type
    FROM likes l
    JOIN comments c ON l.comment_id = c.id
    JOIN posts p ON c.post_id = p.id
    WHERE l.user_id = ? AND l.comment_id IS NOT NULL`, userID)
    if err != nil {
    http.Error(w, "Erreur lors de la récupération des likes des commentaires", http.StatusInternalServerError)
    return
    }

    defer rows.Close()
    
    // // Loop through the fetched likes/dislike comments
    for rows.Next() {
    var commentLike CommentLikeInfo
    if err := rows.Scan(&commentLike.CommentID, &commentLike.Comment, &commentLike.PostTitle, &commentLike.Type); err != nil {
        http.Error(w, "Erreur lors de la lecture des likes des commentaires", http.StatusInternalServerError)
        return
    }
    activity.CommentLikes = append(activity.CommentLikes, commentLike)
    }
	 // Set the response header in JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(activity)
}

// Function to check if the user has the correct role to connect
func RoleMiddleware(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID, userRole, err := GetUserFromSessionRole(r)
        if err != nil {
            fmt.Println("❌ RoleMiddleware: Aucun utilisateur authentifié")
            http.Redirect(w, r, "/", http.StatusFound)
            return
        }
        fmt.Println("👤 Utilisateur:", userID, "| Rôle:", userRole, "| Rôle requis:", requiredRole)

        // Set the role hierarchie
        roleHierarchy := map[string]int{
            "guest":     0,
            "user":      1,
            "moderator": 2,
            "admin":     3,
        }
        userLevel, userExists := roleHierarchy[userRole]
        requiredLevel, requiredExists := roleHierarchy[requiredRole]

        // Check if role exist
        if !userExists || !requiredExists {
            fmt.Println("Rôle inconnu:", userRole, "ou", requiredRole)
            http.Error(w, "Erreur interne", http.StatusInternalServerError)
            return
        }

        // Check if user has a good role or more
        if userLevel < requiredLevel {
            fmt.Println("Accès interdit: Rôle insuffisant")
            http.Error(w, "Accès interdit", http.StatusForbidden)
            return
        }
        next(w, r)
    }
}