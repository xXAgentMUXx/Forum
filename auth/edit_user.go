package auth

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
	"golang.org/x/crypto/bcrypt"
)

// Function to handles editing user information
func EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
        http.ServeFile(w, r, "web/html/edit_user.html")
        return
    }
	// Retrieve the userID from the session to ensure the user is logged in
	userID, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "You need to be logged in to edit your account", http.StatusUnauthorized)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Retrieve form values for username, email, and password
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate the email format
	if !isValidEmail(email) {
		http.Error(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Check if the new email is already taken by another user
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
	// Create the SQL query to update the user information
	updateQuery := "UPDATE users SET username = ?, email = ?"
	args := []interface{}{username, email}


	// If a new password was provided, hash it and add it
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

	// Execute the update
	_, err = DB.Exec(updateQuery, args...)
	if err != nil {
		log.Println("Error updating user:", err)
		http.Error(w, "Error updating user data", http.StatusInternalServerError)
		return
	}
	// Update the session with the new user data
	updateSession(w, r, userID, email)
	fmt.Fprintln(w, "User updated successfully")
}

// Function to updates the session data
func updateSession(w http.ResponseWriter, r *http.Request, userID, email string) {
	// Retrieve the session cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Session not found (not authorized from non-user or user connected from github or Google to modify their account)", http.StatusUnauthorized)
		return
	}
	// Query the database 
	var currentSessionID string
	err = DB.QueryRow("SELECT id FROM sessions WHERE id = ?", cookie.Value).Scan(&currentSessionID)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	// Update the session record with the new user data
	_, err = DB.Exec("UPDATE sessions SET user_id = ?, email = ? WHERE id = ?", userID, email, cookie.Value)

	// Set the session cookie
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

//Function to validates whether the provided email has a valid format
func isValidEmail(email string) bool {
	// Using regex for matching a valid email format
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

// Function to hashes the given password using bcrypt
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
