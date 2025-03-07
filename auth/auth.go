package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

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
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
func LoginUser(w http.ResponseWriter, r *http.Request) {
    if r.Method == http.MethodGet {
        http.ServeFile(w, r, "web/html/login.html")
        return
    }
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    email := r.FormValue("email")
    password := r.FormValue("password")

    var userID, storedPassword string
    err := DB.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    if err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password)); err != nil {
        http.Error(w, "Invalid Password", http.StatusUnauthorized)
        return
    }

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
    if err != nil {
        return "", fmt.Errorf("Aucun cookie de session trouvé")
    }
    var userID string
    err = DB.QueryRow("SELECT user_id FROM sessions WHERE id = ?", cookie.Value).Scan(&userID)
    if err != nil {
        fmt.Println("❌ Session non trouvée en base pour :", cookie.Value)
        return "", fmt.Errorf("Session invalide")
    }
    return userID, nil
}

func CleanupExpiredSessions() {
	DB.Exec("DELETE FROM sessions WHERE expires_at <= CURRENT_TIMESTAMP")
}
func LogoutUser(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_token")
    if err == nil {
        _, err := DB.Exec("DELETE FROM sessions WHERE id = ?", cookie.Value)
        if err != nil {
            http.Error(w, "Erreur lors de la suppression de la session", http.StatusInternalServerError)
            return
        }
    }
    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    "",
        Expires:  time.Now().Add(0 * time.Second),
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    http.Redirect(w, r, "/login", http.StatusFound)
}

func CheckSession(w http.ResponseWriter, r *http.Request) {
    _, err := GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized) 
        return
    }
    w.WriteHeader(http.StatusOK) 
}
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        fmt.Println("AuthMiddleware: Vérification en cours...")

        cookie, err := r.Cookie("session_token")
        if err != nil {
            fmt.Println("❌ Aucun cookie reçu, redirection vers /login")
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        fmt.Println("✅ Cookie reçu :", cookie.Value)

        user, err := GetUserFromSession(r)
        if err != nil {
            fmt.Println("❌ Session invalide, redirection vers /login")
            http.Redirect(w, r, "/login", http.StatusFound)
            return
        }
        fmt.Println("✅ AuthMiddleware: Accès accordé à l'utilisateur", user)
        next(w, r)
    }
}

