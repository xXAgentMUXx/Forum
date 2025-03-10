package auth

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
	"golang.org/x/crypto/bcrypt"
)

func EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
        http.ServeFile(w, r, "web/html/edit_user.html")
        return
    }
	userID, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You need to be logged in to edit your account", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, userID).Scan(&count)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email already taken", http.StatusConflict)
		return
	}
	updateQuery := "UPDATE users SET username = ?, email = ?"
	args := []interface{}{username, email}

	if password != "" {
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}
		updateQuery += ", password = ?"
		args = append(args, hashedPassword)
	}
	updateQuery += " WHERE id = ?"
	args = append(args, userID)

	_, err = DB.Exec(updateQuery, args...)
	if err != nil {
		log.Println("Error updating user:", err)
		http.Error(w, "Error updating user data", http.StatusInternalServerError)
		return
	}
	updateSession(w, r, userID, email)
	fmt.Fprintln(w, "User updated successfully")
}


func updateSession(w http.ResponseWriter, r *http.Request, userID, email string) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Session not found", http.StatusUnauthorized)
		return
	}
	var currentSessionID string
	err = DB.QueryRow("SELECT id FROM sessions WHERE id = ?", cookie.Value).Scan(&currentSessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	_, err = DB.Exec("UPDATE sessions SET user_id = ?, email = ? WHERE id = ?", userID, email, cookie.Value)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    cookie.Value, 
		Expires:  time.Now().Add(24 * time.Hour), 
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	fmt.Println("Session updated successfully")
}


func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
