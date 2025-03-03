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
func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	title := r.FormValue("title")
	content := r.FormValue("content")
	if title == "" || content == "" {
		http.Error(w, "Title and text are required", http.StatusBadRequest)
		return
	}
	postID := uuid.New().String()
	_, err = auth.DB.Exec("INSERT INTO posts (id, user_id, title, content, created_at) VALUES (?, ?, ?, ?, ?)", postID, userID, title, content, time.Now())
	if err != nil {
		http.Error(w, "Error at creating post", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Post created successfully!")
}

func GetPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	var post struct {
		ID        string
		UserID    string
		Title     string
		Content   string
		CreatedAt time.Time
	}
	err := auth.DB.QueryRow("SELECT id, user_id, title, content, created_at FROM posts WHERE id = ?", postID).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(post)
}
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
    filter := r.URL.Query().Get("filter")
    categoryID := r.URL.Query().Get("category_id")
    userID, _ := auth.GetUserFromSession(r) 

    var rows *sql.Rows
    var err error

    if filter == "category" && categoryID != "" {
        rows, err = auth.DB.Query("SELECT id, user_id, title, content, created_at FROM posts WHERE id IN (SELECT post_id FROM post_categories WHERE category_id = ?) ORDER BY created_at DESC", categoryID)
    } else if filter == "my_posts" && userID != "" {
        rows, err = auth.DB.Query("SELECT id, user_id, title, content, created_at FROM posts WHERE user_id = ? ORDER BY created_at DESC", userID)
    } else if filter == "liked" && userID != "" {
        rows, err = auth.DB.Query("SELECT id, user_id, title, content, created_at FROM posts WHERE id IN (SELECT post_id FROM likes WHERE user_id = ? AND type = 'like') ORDER BY created_at DESC", userID)
    } else {
        rows, err = auth.DB.Query("SELECT id, user_id, title, content, created_at FROM posts ORDER BY created_at DESC")
    }
    if err != nil {
        http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type Post struct {
        ID        string    `json:"ID"`
        UserID    string    `json:"UserID"`
        Title     string    `json:"Title"`
        Content   string    `json:"Content"`
        CreatedAt time.Time `json:"CreatedAt"`
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

func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := r.FormValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	var postOwner string
	err = auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}
	if postOwner != userID {
		http.Error(w, "You can only delete your own posts", http.StatusForbidden)
		return
	}
	_, err = auth.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	postID := r.FormValue("post_id")
	typeLike := r.FormValue("type")
	if postID == "" || (typeLike != "like" && typeLike != "dislike") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	likeID := uuid.New().String()
	_, err = auth.DB.Exec("INSERT INTO likes (id, user_id, post_id, type, created_at) VALUES (?, ?, ?, ?, ?)", likeID, userID, postID, typeLike, time.Now())
	if err != nil {
		http.Error(w, "Error liking post", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Like recorded successfully!")
}
func Like_Post(w http.ResponseWriter, r *http.Request) {
	LikeContent(w, r, "post")
}