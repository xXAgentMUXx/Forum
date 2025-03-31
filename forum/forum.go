package forum

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"text/template"
	"time"

	"Forum/auth"

	"github.com/google/uuid"
)

// Function to display the templates for connected user
func ServeForum(w http.ResponseWriter, r *http.Request) {
    userID, err := auth.GetUserFromSession(r)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Récupérer le rôle et l'email de l'utilisateur
    var role, email string
    err = auth.DB.QueryRow("SELECT role, email FROM users WHERE id = ?", userID).Scan(&role, &email)
    if err != nil {
        http.Error(w, "Error retrieving user data", http.StatusInternalServerError)
        return
    }

    // Passer le rôle et l'email au template
    data := struct {
        UserID string
        Role   string
        Email  string
    }{
        UserID: userID,
        Role:   role,
        Email:  email,
    }

    // Charger et exécuter le template
    tmpl, err := template.ParseFiles("web/html/forum.html")
    if err != nil {
        http.Error(w, "Error loading template", http.StatusInternalServerError)
        return
    }

    err = tmpl.Execute(w, data)
    if err != nil {
        http.Error(w, "Error rendering template", http.StatusInternalServerError)
    }
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
	// Create a notification for the owner of the post or comments
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

