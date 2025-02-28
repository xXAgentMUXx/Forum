package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open("sqlite3", "forum.db")
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Database connected!")
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
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&exists)
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
	_, err = db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, string(hashedPassword))
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
	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&userID, &storedPassword)
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
	_, err = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, userID, expiration)
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   sessionID,
		Expires: expiration,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}