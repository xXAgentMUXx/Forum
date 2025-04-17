package forum

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"Forum/auth"
)

// Allowed file extensions for images
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
}

// Maximum allowed size (20mb)
const maxImageSize = 20 * 1024 * 1024 

// Function to save an uploaded image
func saveImage(fileHeader *multipart.FileHeader) (string, error) {
    // Open the uploaded file
	file, err := fileHeader.Open()
    if err != nil {
        return "", err
    }
    defer file.Close()

	// Check the file extension
    ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
    if !allowedExtensions[ext] {
        return "", fmt.Errorf("unsupported file type")
    }
	// Check if the file size exceeds the maximum allowed size
    if fileHeader.Size > maxImageSize {
        return "", fmt.Errorf("file too large")
    }
	// Check if the upload directory exists
    uploadDir := "uploads"
    if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
        err = os.Mkdir(uploadDir, os.ModePerm)
        if err != nil {
            return "", fmt.Errorf("failed to create upload directory: %v", err)
        }
    }
	// Generate a unique ID for the image
    imageID := uuid.New().String()
    filePath := filepath.Join(uploadDir, imageID+ext)

	// Create the file
    outFile, err := os.Create(filePath)
    if err != nil {
        return "", err
    }
    defer outFile.Close()

	// Copy the uploaded file in directory
    _, err = io.Copy(outFile, file)
    if err != nil {
        return "", err
    }
    return filePath, nil
}

// Function to create a new post
func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Get the user ID
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get the form data
	title := r.FormValue("title")
	content := r.FormValue("content")
	categories := r.FormValue("categories") 

	// Check if required fields are provided
	if title == "" || content == "" || categories == "" {
		http.Error(w, "Title, content, and at least one category are required", http.StatusBadRequest)
		return
	}
	// Create a ID for the post
	postID := uuid.New().String()
	var imagePath string

	// Check if an image file is provided
	file, fileHeader, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Save the image and get its path
		imagePath, err = saveImage(fileHeader)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	// Insert the post into the database
	_, err = auth.DB.Exec("INSERT INTO posts (id, user_id, title, content, created_at) VALUES (?, ?, ?, ?, ?)", postID, userID, title, content, time.Now())
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}
	// If an image is provided, insert it into the post_images table
	if imagePath != "" {
		_, err = auth.DB.Exec("INSERT INTO post_images (post_id, image_path) VALUES (?, ?)", postID, imagePath)
		if err != nil {
			http.Error(w, "Error saving image", http.StatusInternalServerError)
			return
		}
	}
	// Add the categories associated with the post
	categoryIDs := strings.Split(categories, ",")
	for _, categoryID := range categoryIDs {
		_, err = auth.DB.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, strings.TrimSpace(categoryID))
		if err != nil {
			http.Error(w, "Error linking post to categories", http.StatusInternalServerError)
			return
		}
	}
	fmt.Fprintf(w, "Post created successfully!")
}

// Function to get a post
func GetPost(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	// Struct for the post
	var post struct {
		ID        string
		UserID    string
		Title     string
		Content   string
		CreatedAt time.Time
		ImagePath string
	}
	// Get the post data
	err := auth.DB.QueryRow("SELECT p.id, p.user_id, p.title, p.content, p.created_at, pi.image_path FROM posts p LEFT JOIN post_images pi ON p.id = pi.post_id WHERE p.id = ?", postID).Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.ImagePath)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}
	// Return the post data in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}


// Function to get all posts with filters
func GetAllPosts(w http.ResponseWriter, r *http.Request) {
    // Get the filter parameters
	filter := r.URL.Query().Get("filter")
    categoryID := r.URL.Query().Get("category_id")
    userID, _ := auth.GetUserFromSession(r)

	// Create the SQL query to retrieve posts
    var rows *sql.Rows
    var err error
    query := `
        SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.created_at, COALESCE(pi.image_path, '')
        FROM posts p
        LEFT JOIN post_images pi ON p.id = pi.post_id
        LEFT JOIN post_categories pc ON p.id = pc.post_id
    `
    if filter == "category" && categoryID != "" {
        query += " WHERE pc.category_id = ? ORDER BY p.created_at DESC"
        rows, err = auth.DB.Query(query, categoryID)
    } else if filter == "my_posts" && userID != "" {
        query += " WHERE p.user_id = ? ORDER BY p.created_at DESC"
        rows, err = auth.DB.Query(query, userID)
    } else if filter == "liked" && userID != "" {
        query += " WHERE p.id IN (SELECT post_id FROM likes WHERE user_id = ? AND type = 'like') ORDER BY p.created_at DESC"
        rows, err = auth.DB.Query(query, userID)
    } else {
        query += " ORDER BY p.created_at DESC"
        rows, err = auth.DB.Query(query)
    }
    if err != nil {
        http.Error(w, "Error retrieving posts", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

	// Define post struct
    type Post struct {
        ID        string    `json:"ID"`
        UserID    string    `json:"UserID"`
        Title     string    `json:"Title"`
        Content   string    `json:"Content"`
        CreatedAt time.Time `json:"CreatedAt"`
        ImagePath string    `json:"ImagePath"`
    }
	// Retrieve the posts
    var posts []Post
    for rows.Next() {
        var post Post
        if err := rows.Scan(&post.ID, &post.UserID, &post.Title, &post.Content, &post.CreatedAt, &post.ImagePath); err != nil {
            http.Error(w, "Error reading post", http.StatusInternalServerError)
            return
        }
        posts = append(posts, post)
    }
	// Return the list of posts in JSON format
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}
// Function to delete a post
func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Get the user ID 
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get the post ID
	postID := r.FormValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	// Get the post owner from the database
	var postOwner string
	err = auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}
	// Ensure the user is the owner of the post
	if postOwner != userID {
		http.Error(w, "You can only delete your own posts", http.StatusForbidden)
		return
	}
	// Delete the post from the database
	_, err = auth.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

// Function to like or dislike a post
func LikePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Get the user ID to ensure the user is authenticated
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Get the post ID and the type of like
	postID := r.FormValue("post_id")
	typeLike := r.FormValue("type")

	// Validate the input
	if postID == "" || (typeLike != "like" && typeLike != "dislike") {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}
	// Generate a ID for the like action
	likeID := uuid.New().String()

	// Insert the like or dislike into the database
	_, err = auth.DB.Exec("INSERT INTO likes (id, user_id, post_id, type, created_at) VALUES (?, ?, ?, ?, ?)", likeID, userID, postID, typeLike, time.Now())
	if err != nil {
		http.Error(w, "Error liking post", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Like recorded successfully!")
}

// Function to invoke the LikeContent function for a post
func Like_Post(w http.ResponseWriter, r *http.Request) {
	LikeContent(w, r, "post")
}