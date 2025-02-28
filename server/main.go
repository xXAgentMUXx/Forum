package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_"github.com/mattn/go-sqlite3"
	"github.com/gorilla/mux"
	"Forum/auth"
)

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	
	r := mux.NewRouter()

	r.HandleFunc("/register", RegisterHandler).Methods("POST")
	r.HandleFunc("/login", LoginHandler).Methods("POST")

	fmt.Println("Serveur démarré sur :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user auth.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Données invalides", http.StatusBadRequest)
		return
	}
	err = auth.RegisterUser(db, user.Email, user.Username, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Register Complete"})
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Email    string `json:"E-mail"`
		Password string `json:"Password"`
	}
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalides Data", http.StatusBadRequest)
		return
	}
	err = auth.LoginUser(db, creds.Email, creds.Password, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
}
