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
	mux := http.NewServeMux()
	mux.HandleFunc("/forum", Forum.Forum)
}