package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"Forum/auth"
)
func CreateComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	postID := r.FormValue("post_id")
	content := r.FormValue("content")
	if postID == "" || content == "" {
		http.Error(w, "Post ID and text are required", http.StatusBadRequest)
		return
	}
	commentID := uuid.New().String()
	_, err = auth.DB.Exec("INSERT INTO comments (id, user_id, post_id, content, created_at) VALUES (?, ?, ?, ?, ?)", commentID, userID, postID, content, time.Now())
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Comment created successfully!")
}
func GetComments(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("post_id")
    if postID == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }

    rows, err := auth.DB.Query("SELECT id, user_id, content, created_at FROM comments WHERE post_id = ? ORDER BY created_at ASC", postID)
    if err != nil {
        http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type Comment struct {
        ID        string    `json:"id"`
        UserID    string    `json:"user_id"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
    }
    var comments []Comment
    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
            http.Error(w, "Error at reading comment", http.StatusInternalServerError)
            return
        }
        comments = append(comments, comment)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	commentID := r.FormValue("id")
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}
	var commentOwner string
	err = auth.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving comment", http.StatusInternalServerError)
		return
	}
	if commentOwner != userID {
		http.Error(w, "You can only delete your own comments", http.StatusForbidden)
		return
	}
	_, err = auth.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		http.Error(w, "error deleting comment", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "comment deleted successfully!")
}
func LikeComment(w http.ResponseWriter, r *http.Request) {
	LikeContent(w, r, "comment")
}