package forum

import (
	"encoding/json"
	"net/http"
	"time"

	"Forum/auth"
)
func GetFilteredPosts(w http.ResponseWriter, r *http.Request) {
    filterType := r.URL.Query().Get("filter")
    userID, _ := auth.GetUserFromSession(r) 

    var query string
    var args []interface{}

    switch filterType {
    case "category":
        categoryID := r.URL.Query().Get("category_id")
        if categoryID == "" {
            http.Error(w, "Category ID is required", http.StatusBadRequest)
            return
        }
        query = `SELECT p.id, p.user_id, p.title, p.content, p.created_at FROM posts p
                 JOIN post_categories pc ON p.id = pc.post_id
                 WHERE pc.category_id = ? ORDER BY p.created_at DESC`
        args = append(args, categoryID)

    case "my_posts":
        if userID == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        query = "SELECT id, user_id, title, content, created_at FROM posts WHERE user_id = ? ORDER BY created_at DESC"
        args = append(args, userID)

    case "liked":
        if userID == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        query = `SELECT p.id, p.user_id, p.title, p.content, p.created_at FROM posts p
                 JOIN likes l ON p.id = l.post_id
                 WHERE l.user_id = ? AND l.type = 'like' ORDER BY p.created_at DESC`
        args = append(args, userID)

    default:
        query = "SELECT id, user_id, title, content, created_at FROM posts ORDER BY created_at DESC"
    }
    rows, err := auth.DB.Query(query, args...)
    if err != nil {
        http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type Post struct {
        ID        string    `json:"id"`
        UserID    string    `json:"user_id"`
        Title     string    `json:"title"`
        Content   string    `json:"content"`
        CreatedAt time.Time `json:"created_at"`
    }
    var posts []Post

    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt); err != nil {
            http.Error(w, "Error reading post", http.StatusInternalServerError)
            return
        }
        posts = append(posts, post)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
