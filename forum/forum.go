package forum

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"Forum/auth"

	"github.com/google/uuid"
)

// Function to display the templates for connected user
func ServeForum(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/forum.html")
}

// Function to display the templates for non-connected user
func ServeForumInvite(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/forum_invite.html")
}

// Function to handles liking or disliking a post or comment.
func LikeContent(w http.ResponseWriter, r *http.Request, contentType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Get the user ID to verify if the user is logged in.
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Retrieve the content ID
	contentID := r.FormValue("id")
	typeLike := r.FormValue("type")

	// Validate the input
	if contentID == "" || (typeLike != "like" && typeLike != "dislike") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	// Check if the user has already liked or disliked
	var existingType string
	err = auth.DB.QueryRow("SELECT type FROM likes WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", userID, contentID, contentID).Scan(&existingType)

	if err == sql.ErrNoRows {
		// If no like/dislike exists, create a new like/dislike entry in the database
		likeID := uuid.New().String()
		var query string

		// Determine the query based on the content type
		if contentType == "post" {
			query = "INSERT INTO likes (id, user_id, post_id, type, created_at) VALUES (?, ?, ?, ?, ?)"
		} else {
			query = "INSERT INTO likes (id, user_id, comment_id, type, created_at) VALUES (?, ?, ?, ?, ?)"
		}

		// Insert the new like/dislike into the database.
		_, err = auth.DB.Exec(query, likeID, userID, contentID, typeLike, time.Now())
	} else if err == nil {
		// If the user has already liked or disliked, update the existing record
		if existingType == typeLike {
			_, err = auth.DB.Exec("DELETE FROM likes WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", userID, contentID, contentID)
		} else {
			// If the user wants to change their like/dislike, update the record
			_, err = auth.DB.Exec("UPDATE likes SET type = ? WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", typeLike, userID, contentID, contentID)
		}
	} else {
		http.Error(w, "Error processing like", http.StatusInternalServerError)
		return
	}
	var ownerID string
	if contentType == "post" {
		err = auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", contentID).Scan(&ownerID)
	} else {
		err = auth.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", contentID).Scan(&ownerID)
	}
	if err == nil && ownerID != userID {
		CreateNotification(ownerID, contentID, "like", "Your post/comment received a "+typeLike)
	}
	if err != nil {
		http.Error(w, "Error updating like status", http.StatusInternalServerError)
		return
	}
	if err == nil && ownerID != userID {
		CreateNotification(ownerID, contentID, "like", "Your post/comment received a "+typeLike)
	}
	// Send a JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Like status updated successfully"})
}

// Function to retrieves the count of likes and dislikes
func GetLikesAndDislike(w http.ResponseWriter, r *http.Request) {
	// Get content ID and type
	contentID := r.URL.Query().Get("id")
	contentType := r.URL.Query().Get("type")

	// Validate the content ID and type parameters
	if contentID == "" || (contentType != "post" && contentType != "comment") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	var likeCount, dislikeCount int
	var query string

	// Determine the query based on the content type
	if contentType == "post" {
		query = "SELECT COALESCE(COUNT(CASE WHEN type='like' THEN 1 END), 0), COALESCE(COUNT(CASE WHEN type='dislike' THEN 1 END), 0) FROM likes WHERE post_id = ?"
	} else {
		query = "SELECT COALESCE(COUNT(CASE WHEN type='like' THEN 1 END), 0), COALESCE(COUNT(CASE WHEN type='dislike' THEN 1 END), 0) FROM likes WHERE comment_id = ?"
	}
	// Execute the query to get the like and dislike counts
	err := auth.DB.QueryRow(query, contentID).Scan(&likeCount, &dislikeCount)
	if err != nil {
		http.Error(w, "Error retrieving like count", http.StatusInternalServerError)
		return
	}
	// Return the like and dislike counts in a JSON response.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": likeCount, "dislikes": dislikeCount})
}

// Function to retrieves all categories from the database
func GetCategories(w http.ResponseWriter, r *http.Request) {
	// Query the database to get all categories.
	rows, err := auth.DB.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Define a category struct
	type Category struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	// Create a slice to hold the categories
	var categories []Category
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			http.Error(w, "Error reading categories", http.StatusInternalServerError)
			return
		}
		// Add each category to the slice.
		categories = append(categories, category)
	}
	//Return a JSON respons
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func CreateNotification(userID, postID, action, content string) {
	notificationID := uuid.New().String()
	_, err := auth.DB.Exec("INSERT INTO notifications (id, user_id, post_id, action, content, created_at) VALUES (?, ?, ?, ?, ?, ?)", notificationID, userID, postID, action, content, time.Now())
	if err != nil {
		return
	}
}

func GetNotifications(w http.ResponseWriter, r *http.Request) {

    userID, err := auth.GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    rows, err := auth.DB.Query("SELECT n.id, n.post_id, n.action, n.content, n.created_at, n.seen, u.username FROM notifications n JOIN users u ON n.user_id = u.id WHERE n.user_id = ? ORDER BY n.created_at DESC", userID)

    if err != nil {
        http.Error(w, "Error retrieving notifications", http.StatusInternalServerError)
        return
    }
    defer rows.Close()
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

    for rows.Next() {
        var notif Notification
        if err := rows.Scan(&notif.ID, &notif.PostID, &notif.Action, &notif.Content, &notif.CreatedAt, &notif.Seen, &notif.Username); err != nil {
            http.Error(w, "Error reading notifications", http.StatusInternalServerError)
            return
        }
        
        // Récupérer l'acteur de la notification
        var actorID string
        var query string
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
    w.Header().Set("Content-Type", "application/json")

    json.NewEncoder(w).Encode(notifications)

}

func MarkNotificationsAsSeen(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err = auth.DB.Exec("UPDATE notifications SET seen = true WHERE user_id = ?", userID)
	if err != nil {
		http.Error(w, "Error updating notifications", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func GetNewComments(w http.ResponseWriter, r *http.Request) {
    postID := r.URL.Query().Get("post_id")
    if postID == "" {
        http.Error(w, "Post ID is required", http.StatusBadRequest)
        return
    }
    rows, err := auth.DB.Query("SELECT id, user_id, content, created_at FROM comments WHERE post_id = ? ORDER BY created_at DESC", postID)
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
            http.Error(w, "Error reading comments", http.StatusInternalServerError)
            return
        }
        comments = append(comments, comment)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(comments)
}

func DeleteNotification(w http.ResponseWriter, r *http.Request) {
    userID, err := auth.GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    notifID := r.FormValue("id")
    if notifID == "" {
        http.Error(w, "Notification ID is required", http.StatusBadRequest)
        return
    }
    _, err = auth.DB.Exec("DELETE FROM notifications WHERE id = ? AND user_id = ?", notifID, userID)
    if err != nil {
        http.Error(w, "Error deleting notification", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}