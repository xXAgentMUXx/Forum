package auth

import (
	"database/sql"
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
)

type User struct {
	ID       string
	Email    string
	Username string
	Password string
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func RegisterUser(db *sql.DB, email, username, password string) error {
	var exists bool
	err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM users WHERE email = ? OR username = ?)", email, username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email ou username déjà utilisé")
	}
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return err
	}
	userID := uuid.New().String()

	_, err = db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, hashedPassword)
	if err != nil {
    return err
	}	
	log.Println("Utilisateur enregistré:", userID, email, username)
	return nil
}

/* curl -X POST http://localhost:8080/register \
     -H "Content-Type: application/json" \
     -d '{"Email": "test@example.com", "Password": "password123"}'*/
