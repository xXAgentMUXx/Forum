package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
    
	security "Forum/security"
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
	)`);
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
            fmt.Println("✅ [DEBUG] Utilisateur trouvé via session_token:", userID)
            return userID, nil
        }
    }
    session, _ := store.Get(r, "session-name")
    email, ok := session.Values["email"].(string)
    fmt.Println("🛠️ [DEBUG] Email trouvé en session OAuth :", email)

    if ok && email != "" {
        var userID string
        err := DB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
        
        if err == sql.ErrNoRows {
            fmt.Println("🛠️ [DEBUG] Création d'un nouvel utilisateur OAuth :", email)
            userID = uuid.New().String()
            _, err = DB.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, NULL)", userID, email, email)
            if err != nil {
                fmt.Println("❌ [DEBUG] Erreur lors de la création de l'utilisateur OAuth :", err)
                return "", fmt.Errorf("Erreur lors de la création de l'utilisateur OAuth")
            }
            fmt.Println("✅ [DEBUG] Utilisateur OAuth ajouté en base :", userID)
        } else if err != nil {
            return "", fmt.Errorf("Erreur de base de données")
        }
        return userID, nil
    }
    fmt.Println("❌ [DEBUG] Aucune session valide trouvée")
    return "", fmt.Errorf("Aucune session valide trouvée")
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
        Expires:  time.Now().Add(-1 * time.Hour), 
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    })
    session, _ := store.Get(r, "session-name")
    session.Values["email"] = nil
    session.Options.MaxAge = -1 
    err = session.Save(r, w)
    if err != nil {
        fmt.Println("❌ Erreur lors de la suppression de la session OAuth:", err)
    }
    http.Redirect(w, r, "/login", http.StatusFound)
}

func CheckSession(w http.ResponseWriter, r *http.Request) {
    userID, err := GetUserFromSession(r)
    if err == nil {
        fmt.Println("✅ CheckSession: Utilisateur connecté avec session classique", userID)
        w.WriteHeader(http.StatusOK)
        return
    }
    session, _ := store.Get(r, "session-name")
    if email, ok := session.Values["email"].(string); ok && email != "" {
        fmt.Println("✅ CheckSession: Utilisateur connecté via OAuth", email)
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
        session, _ := store.Get(r, "session-name")
        fmt.Println("🛠️ Contenu de la session:", session.Values)

        if email, ok := session.Values["email"].(string); ok && email != "" {
            fmt.Println("✅ AuthMiddleware: Accès accordé à l'utilisateur OAuth", email)
            next(w, r)
            return
        }
        fmt.Println("❌ AuthMiddleware: Aucun utilisateur authentifié, redirection vers /login")
        http.Redirect(w, r, "/login", http.StatusFound)
    }
}


