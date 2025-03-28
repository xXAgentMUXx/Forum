package forum

import (
	"Forum/auth"
	"encoding/json"
	"net/http"
	"time"
	"github.com/google/uuid"
)

// Function to creates a new notification for a user related to a post
func CreateNotification(userID, postID, action, content string) {
	notificationID := uuid.New().String()
	_, err := auth.DB.Exec("INSERT INTO notifications (id, user_id, post_id, action, content, created_at) VALUES (?, ?, ?, ?, ?, ?)", notificationID, userID, postID, action, content, time.Now())
	if err != nil {
		return
	}
}

// Function to retrieves a list of notifications for a user
func GetNotifications(w http.ResponseWriter, r *http.Request) {
	// Get the user ID
    userID, err := auth.GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
	// Query the notifications from the database
    rows, err := auth.DB.Query("SELECT n.id, n.post_id, n.action, n.content, n.created_at, n.seen, u.username FROM notifications n JOIN users u ON n.user_id = u.id WHERE n.user_id = ? ORDER BY n.created_at DESC", userID)

    if err != nil {
        http.Error(w, "Error retrieving notifications", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

	// Define a struct for a notification
    type Notification struct {
        ID        string    `json:"id"`
        PostID    string    `json:"post_id"`
        Action    string    `json:"action"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
        Seen      bool      `json:"seen"`
        Username  string    `json:"username"`
    }
    var notifications []Notification

	// Loop through the rows and create a list of notifications
    for rows.Next() {
        var notif Notification
        if err := rows.Scan(&notif.ID, &notif.PostID, &notif.Action, &notif.Content, &notif.CreatedAt, &notif.Seen, &notif.Username); err != nil {
            http.Error(w, "Error reading notifications", http.StatusInternalServerError)
            return
        }
        
        // Récupérer l'acteur de la notification
        var actorID string
        var query string

		// Retrieve the actor of the notification based on the action (like or comment)
        if notif.Action == "comment" {
            query = "SELECT user_id FROM comments WHERE post_id = ? ORDER BY created_at DESC LIMIT 1"
        } else {
            query = "SELECT user_id FROM likes WHERE post_id = ? AND type = 'like' ORDER BY created_at DESC LIMIT 1"
        }
        if err := auth.DB.QueryRow(query, notif.PostID).Scan(&actorID); err == nil {
            auth.DB.QueryRow("SELECT username FROM users WHERE id = ?", actorID).Scan(&notif.Username)
        }
        notifications = append(notifications, notif)

    }
    // Set the response header to JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(notifications)
}

// Function to marks all notifications as "seen" for a user
func MarkNotificationsAsSeen(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Update the "seen" status for the notifications
	_, err = auth.DB.Exec("UPDATE notifications SET seen = true WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Error updating notifications", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// function to retrieves the most recent comments for a specific post
func GetNewComments(w http.ResponseWriter, r *http.Request) {
	// Get the post ID 
    postID := r.URL.Query().Get("post_id")
    if postID == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }
	// Query the comments from the database
    rows, err := auth.DB.Query("SELECT id, user_id, content, created_at FROM comments WHERE post_id = ? ORDER BY created_at DESC", postID)
    if err != nil {
        http.Error(w, "Error retrieving comments", http.StatusInternalServerError)
        return
    }
	// Define a struct for a comment
    defer rows.Close()
    type Comment struct {
        ID        string    `json:"id"`
        UserID    string    `json:"user_id"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
    }
	// Loop through the rows and add comments to the list
    var comments []Comment
    for rows.Next() {
        var comment Comment
        if err := rows.Scan(&comment.ID, &comment.UserID, &comment.Content, &comment.CreatedAt); err != nil {
            http.Error(w, "Error reading comments", http.StatusInternalServerError)
            return
        }
        comments = append(comments, comment)
    }
	// Set the response header to JSON
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

// Function to deletes a specific notification for a user
func DeleteNotification(w http.ResponseWriter, r *http.Request) {
	// Get the user ID 
    userID, err := auth.GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
	// Get the notification ID
    notifID := r.FormValue("id")
    if notifID == "" {
        http.Error(w, "Notification ID is required", http.StatusBadRequest)
        return
    }
	// Delete the notification from the database
    _, err = auth.DB.Exec("DELETE FROM notifications WHERE id = ? AND user_id = ?", notifID, userID)
    if err != nil {
        http.Error(w, "Error deleting notification", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}