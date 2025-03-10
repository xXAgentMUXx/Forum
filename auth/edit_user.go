package auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func init() {
    var err error
    db, err = sql.Open("sqlite3", "./forum.db")
    if err != nil {
        log.Fatal(err)
    }
}

func ServeEditUserPage(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "web/html/edit_user.html")
}

func EditUser(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "user-session")
    userID, ok := session.Values["user_id"].(string)
    if !ok || userID == "" {
        http.Error(w, "You need to be logged in to edit your account", http.StatusUnauthorized)
        return
    }
    if r.Method == http.MethodPost {
        username := r.FormValue("username")
        email := r.FormValue("email")
        password := r.FormValue("password")

        if !isValidEmail(email) {
            http.Error(w, "Invalid email format", http.StatusBadRequest)
            return
        }
        var existingEmail string
        err := db.QueryRow("SELECT email FROM users WHERE email = ? AND id != ?", email, userID).Scan(&existingEmail)
        if err != nil && err != sql.ErrNoRows {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        if existingEmail != "" {
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

        _, err = db.Exec(updateQuery, args...)
        if err != nil {
            http.Error(w, "Error updating user data", http.StatusInternalServerError)
            return
        }
        fmt.Fprintln(w, "User updated successfully")
    }
}
func updateUser(w http.ResponseWriter, r *http.Request) {
    session, _ := store.Get(r, "session-name")
	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
    http.Error(w, "You need to be logged in to edit your account", http.StatusUnauthorized)
    return
	}
    
    // Récupérer les nouvelles informations
    newEmail := r.FormValue("email")
    newPassword := r.FormValue("password")

	updateQuery := "UPDATE users SET"
	args := []interface{}{}
	
	if newEmail != "" {
		updateQuery += " email = ?,"
		args = append(args, newEmail)
	}
	
	if newPassword != "" {
		hashedPassword, err := hashPassword(newPassword)
		if err != nil {
			http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
			return
		}
		updateQuery += " password = ?,"
		args = append(args, hashedPassword)
	}
	
	if len(args) == 0 {
		http.Error(w, "Aucune modification effectuée", http.StatusBadRequest)
		return
	}
	
	// Retirer la dernière virgule et ajouter la condition WHERE
	updateQuery = updateQuery[:len(updateQuery)-1] + " WHERE id = ?"
	args = append(args, userID)
	
	_, err := db.Exec(updateQuery, args...)
	if err != nil {
		log.Println("Erreur de mise à jour utilisateur:", err)
		http.Error(w, "Erreur de mise à jour de l'utilisateur", http.StatusInternalServerError)
		return
	}
    session.Values["user_id"] = userID
	session.Values["email"] = newEmail
	session.Save(r, w)
    fmt.Fprintf(w, "Informations mises à jour avec succès!")
}

// Assurez-vous que la session est bien mise à jour
func HandleEditUser(w http.ResponseWriter, r *http.Request) {
    // Si la méthode est GET, afficher le formulaire
    if r.Method == http.MethodGet {
        ServeEditUserPage(w, r)  // Afficher la page de modification
        return
    }

    // Si la méthode est POST, traiter la soumission du formulaire
    if r.Method == http.MethodPost {
        updateUser(w, r)  // Mettre à jour l'utilisateur dans la base de données
    }
}


func isValidEmail(email string) bool {
    re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
    return re.MatchString(email)
}

func hashPassword(password string) (string, error) {
    hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hashed), err
}

func checkPasswordHash(password, hashedPassword string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}
func updateSession(w http.ResponseWriter, r *http.Request, userID string) {
    session, _ := store.Get(r, "session-name") 
    session.Values["userID"] = userID          
    session.Save(r, w)                         
}