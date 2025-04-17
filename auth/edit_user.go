package auth

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"
	"html/template"
	"golang.org/x/crypto/bcrypt"
)

var tmpl = template.Must(template.ParseFiles("web/html/edit_user.html"))

// Function to handles editing user information
func EditUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Retreives the user ID session to ckeck if he is connected
		userID, err := GetUserFromSession(r)
		if err != nil {
			http.Error(w, "Vous devez être connecté pour modifier votre compte", http.StatusUnauthorized)
			return
		}

		// Retrieves role user from database
		var role string
		err = DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération du rôle de l'utilisateur", http.StatusInternalServerError)
			return
		}
		// Past the ID for the role and email of the user
		tmplData := struct {
			UserID string
			Role   string
		}{
			UserID: userID,
			Role:   role,
		}
		// Served the template with the role and email
		err = tmpl.Execute(w, tmplData)
		if err != nil {
			log.Println("Erreur lors du rendu du template:", err)
			http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		}
		return
	}
	// Check if methode is GET
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode de requête non valide", http.StatusMethodNotAllowed)
		return
	}
	// Retrieves ID user from sesssion
	userID, err := GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Vous devez être connecté pour modifier votre compte", http.StatusUnauthorized)
		return
	}
	// Retrieves the informations necessary
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate the email of the user
	if !isValidEmail(email) {
		http.Error(w, "Format d'email invalide", http.StatusBadRequest)
		return
	}
	// Check if email is not already taken for another user
	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? AND id != ?", email, userID).Scan(&count)
	if err != nil {
		http.Error(w, "Erreur de base de données", http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "L'email est déjà pris", http.StatusConflict)
		return
	}
	// Create the SQL request for update the user
	updateQuery := "UPDATE users SET username = ?, email = ?"
	args := []interface{}{username, email}

	// If new password is created, then it is hased for security
	if password != "" {
		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
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
		log.Println("Erreur lors de la mise à jour de l'utilisateur:", err)
		http.Error(w, "Erreur lors de la mise à jour des données utilisateur", http.StatusInternalServerError)
		return
	}
	// Update the session with new informations of the user
	updateSession(w, r, userID, email)
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
	http.Redirect(w, r, "/login", http.StatusSeeOther)
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