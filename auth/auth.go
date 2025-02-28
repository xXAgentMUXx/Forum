package auth

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string
	Email    string
	Username string
	Password string
}

func EnterPassword(password string) (string, error) {
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
		return fmt.Errorf("E-mail or username already used")
	}
	hashedPassword, err := EnterPassword(password)
	if err != nil {
		return err
	}
	userID := uuid.New().String()
	
	_, err = db.Exec("INSERT INTO users (id, email, username, password) VALUES (?, ?, ?, ?)", userID, email, username, hashedPassword)
	if err != nil {
		return err
	}
	log.Println("User registred with sucess")
	return nil
}
