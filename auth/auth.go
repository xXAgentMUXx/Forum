package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
	"Forum/security"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var DB *sql.DB

func InitDB() {
	var err error
	DB, err = sql.Open("sqlite3", "forum.db")
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
}

func ServeHTML(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/"+r.URL.Path)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/html/register.html")
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")

	var exists string
	err := DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&exists)
	if err == nil {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error encrypting password", http.StatusInternalServerError)
		return
	}
	userID := uuid.New().String()
	_, err = DB.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error registering user (pseudo already used)", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

var loginLimiter = security.NewLoginLimiter()

func LoginUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "web/html/login.html")
		return
	}
	ip := r.RemoteAddr
	if blocked, remaining := loginLimiter.CheckLock(ip); blocked {
		http.Error(w, fmt.Sprintf("Trop de tentatives. Réessayez dans %v secondes.", int(remaining.Seconds())), http.StatusTooManyRequests)
		return
	}
	email := r.FormValue("email")
	password := r.FormValue("password")

	var userID, storedPassword string
	err := DB.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
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
	if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
		time.Sleep(4 * time.Second)
		timeout := loginLimiter.FailedAttempt(ip)
		if timeout > 0 {
			http.Error(w, fmt.Sprintf("Trop de tentatives. Réessayez dans %v secondes.", int(timeout.Seconds())), http.StatusTooManyRequests)
		} else {
			http.Error(w, "Invalid Password", http.StatusUnauthorized)
		}
		return
	}
	loginLimiter.Reset(ip)
	DB.Exec("DELETE FROM rate_limit WHERE user_id = ?", userID)

	sessionID := uuid.New().String()
	expiration := time.Now().Add(24 * time.Hour)
	_, err = DB.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiration)
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
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/forum", http.StatusSeeOther)
}

func GetUserFromSession(r *http.Request) (string, error) {
    cookie, err := r.Cookie("session_token")
    if err == nil {
        var userID string
        err = DB.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)
        if err == nil {
            return userID, nil
        }
    }

    
    oauthCookie, err := r.Cookie("session")
    if err == nil && oauthCookie.Value != "" {
        return oauthCookie.Value, nil 
    }

    return "", fmt.Errorf("No valid session found")
}

func CleanupExpiredSessions() {
	DB.Exec("DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP")
}

func LogoutUser(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err == nil {
        DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
    }
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

    fmt.Println("✅ Déconnexion réussie. Redirection vers /login")
    http.Redirect(w, r, "/login", http.StatusFound)
}


func CheckSession(w http.ResponseWriter, r *http.Request) {
    userID, err := GetUserFromSession(r)
    if err == nil {
        fmt.Println("✅ CheckSession: Utilisateur connecté:", userID)
        w.WriteHeader(http.StatusOK)
        return
    }
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("🔍 AuthMiddleware: Vérification en cours...")
        userID, err := GetUserFromSession(r)
        if err == nil {
            fmt.Println("✅ AuthMiddleware: Accès accordé à l'utilisateur", userID)
            next(w, r)
            return
        }
        fmt.Println("❌ AuthMiddleware: Aucun utilisateur authentifié, redirection vers /login")
        http.Redirect(w, r, "/login", http.StatusFound)
    }
}

