package forum

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"Forum/auth"

	"github.com/google/uuid"
)
func ServeForum(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/forum.html")
}
func ServeForumInvite(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/forum_invite.html")
}
func LikeContent(w http.ResponseWriter, r *http.Request, contentType string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	contentID := r.FormValue("id")
	typeLike := r.FormValue("type")
	if contentID == "" || (typeLike != "like" && typeLike != "dislike") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	var existingType string
	err = auth.DB.QueryRow("SELECT type FROM likes WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", userID, contentID, contentID).Scan(&existingType)

	if err == sql.ErrNoRows {
		likeID := uuid.New().String()
		var query string
		if contentType == "post" {
			query = "INSERT INTO likes (id, user_id, post_id, type, created_at) VALUES (?, ?, ?, ?, ?)"
		} else {
			query = "INSERT INTO likes (id, user_id, comment_id, type, created_at) VALUES (?, ?, ?, ?, ?)"
		}
		_, err = auth.DB.Exec(query, likeID, userID, contentID, typeLike, time.Now())
	} else if err == nil {
		if existingType == typeLike {
			_, err = auth.DB.Exec("DELETE FROM likes WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", userID, contentID, contentID)
		} else {
			_, err = auth.DB.Exec("UPDATE likes SET type = ? WHERE user_id = ? AND (post_id = ? OR comment_id = ?)", typeLike, userID, contentID, contentID)
		}
	} else {
		http.Error(w, "Error processing like", http.StatusInternalServerError)
		return
	}
	if err != nil {
		http.Error(w, "Error updating like status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Like status updated successfully"})
}
func GetLikesAndDislike(w http.ResponseWriter, r *http.Request) {
	contentID := r.URL.Query().Get("id")
	contentType := r.URL.Query().Get("type")

	if contentID == "" || (contentType != "post" && contentType != "comment") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	var likeCount, dislikeCount int
	var query string

	if contentType == "post" {
		query = "SELECT COALESCE(COUNT(CASE WHEN type='like' THEN 1 END), 0), COALESCE(COUNT(CASE WHEN type='dislike' THEN 1 END), 0) FROM likes WHERE post_id = ?"
	} else {
		query = "SELECT COALESCE(COUNT(CASE WHEN type='like' THEN 1 END), 0), COALESCE(COUNT(CASE WHEN type='dislike' THEN 1 END), 0) FROM likes WHERE comment_id = ?"
	}
	err := auth.DB.QueryRow(query, contentID).Scan(&likeCount, &dislikeCount)
	if err != nil {
		http.Error(w, "Error retrieving like count", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"likes": likeCount, "dislikes": dislikeCount})
}


