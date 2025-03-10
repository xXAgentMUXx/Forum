package auth

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
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
	session, _ := store.Get(r, "session-name")
	session.Values["userID"] = userID
	session.Values["email"] = email
	session.Save(r, w)
}

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
