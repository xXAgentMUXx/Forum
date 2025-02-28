package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)


func LoginUser(db *sql.DB, email, password string, w http.ResponseWriter) error {
	var user User

	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("email non trouvé")
		}
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return errors.New("mot de passe incorrect")
	}
	
	sessionID := uuid.New().String()

	_, err = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, user.ID, time.Now().Add(24*time.Hour))
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "session",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	})
	return nil
}
/*curl -X POST http://localhost:8080/login \
     -H "Content-Type: application/json" \
     -d '{"Email": "test@gmail.com", "Password": "password123"}'*/