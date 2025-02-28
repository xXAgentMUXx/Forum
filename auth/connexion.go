package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
		return errors.New("Incorrect Password")
	}
	sessionID := uuid.New().String()

	_, err = db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)", sessionID, user.ID, time.Now().Add(24*time.Hour))
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "cookie",
		Value:   sessionID,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	})
	return nil
}