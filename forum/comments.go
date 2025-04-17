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
// Function for the creation of a new comment
func CreateComment(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := r.FormValue("post_id")
	content := r.FormValue("content")

	commentID := uuid.New().String()
	_, err = auth.DB.Exec("INSERT INTO comments (id, user_id, post_id, content, created_at) VALUES (?, ?, ?, ?, ?)", commentID, userID, postID, content, time.Now())
	if err != nil {
		http.Error(w, "Error creating comment", http.StatusInternalServerError)
		return
	}
	var postOwner string
	err = auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwner)

	// Create a notification for the owner of the post
	if err == nil && postOwner != userID {
		CreateNotification(postOwner, postID, "comment", content)
	}
}

// Function to retrieves all comments associated with a specific post
func GetComments(w http.ResponseWriter, r *http.Request) {
	// Get the post ID
    postID := r.URL.Query().Get("post_id")

	// Query the database for comments
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

	// Define a struct to hold the comment data
    type Comment struct {
        ID        string    `json:"id"`
        UserID    string    `json:"user_id"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
    }
	// Initialize a slice to store the comments
    var comments []Comment

	// Iterate through the rows and scan the data
    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
            http.Error(w, "Error at reading comment", http.StatusInternalServerError)
            return
        }
		// Append the comment to the slice
        comments = append(comments, comment)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

//Function to handles the deletion of a comment.
func DeleteComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Retrieve the user ID from the session to ensure the user is authenticated
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	// Get the comment ID
	commentID := r.FormValue("id")
	if commentID == "" {
		http.Error(w, "comment ID is required", http.StatusBadRequest)
		return
	}
	// Query the database
	var commentOwner string
	err = auth.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "comment not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving comment", http.StatusInternalServerError)
		return
	}
	// Check if the current user is the owner of the comment
	if commentOwner != userID {
		http.Error(w, "You can only delete your own comments", http.StatusForbidden)
		return
	}
	// Delete the comment from the database
	_, err = auth.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		http.Error(w, "error deleting comment", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "comment deleted successfully!")
}

// Function to handles liking a comment
func LikeComment(w http.ResponseWriter, r *http.Request) {
	LikeContent(w, r, "comment")
}