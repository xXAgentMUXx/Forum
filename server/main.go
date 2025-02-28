package main

import (
	"Forum"
	"net/http"
	"time"
)
type Index struct {
	ThreadSubject   string
	ThreadCreatedAt time.Time
	ThreadUUID      string
	Username        string
	PostCount       int
}

func main() {
	mux := http.NewServeMux() // mux sert pour se connecter a son compte google
	mux.HandleFunc("/forum", Forum.Forum)
}