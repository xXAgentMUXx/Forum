package forum

import (
	"Forum/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// function to display the templates moderator
func ServeModerator(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/moderator.html")
}
// function to display the templates admin
func ServeAdmin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/admin.html")
}

// Function to deletes a post by the admin
func DeletePostByAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Retrieve the post ID to delete
	postID := r.FormValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}
	// Check if the post exists in the database
	var postOwner string
	err := auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}
	// Delete the post from the database
	_, err = auth.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}
	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

// DeleteCommentAdmin deletes a comment by the admin
func DeleteCommentAdmin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    // Retrieve the comment ID
    commentID := r.FormValue("id")
    if commentID == "" {
        http.Error(w, "Comment ID is required", http.StatusBadRequest)
        return
    }
    // Check if the comment exists in the database
    var commentOwner string
    err := auth.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwner)
    if err == sql.ErrNoRows {
        http.Error(w, "Comment not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Error retrieving comment", http.StatusInternalServerError)
        return
    }
    // Delete the comment from the database
    _, err = auth.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
    if err != nil {
        http.Error(w, "Error deleting comment", http.StatusInternalServerError)
        return
    }
}

// Function to allows users to report posts
func ReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	// Check if the database connection is active
	if auth.DB == nil {
		log.Println("Database connection is nil")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	// Retrieve the form data
	postID := r.FormValue("id")
	reason := r.FormValue("reason")

	if postID == "" || reason == "" {
		log.Println("Missing parameters: postID or reason")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}
	// Insert the report into the database
	query := "INSERT INTO reports (post_id, reason, status) VALUES (?, ?, 'pending')"
	log.Println("Executing SQL Query:", query)

	_, err := auth.DB.Exec(query, postID, reason)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Error creating report", http.StatusInternalServerError)
		return
	}
	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Report submitted successfully"})
}

// Function to allows the admin to resolve a report
func ResolveReport(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    reportID := r.FormValue("id")
    if reportID == "" {
        http.Error(w, "Report ID is required", http.StatusBadRequest)
        return
    }
    // Update the report status in the database
    _, err := auth.DB.Exec("UPDATE reports SET status = 'resolved' WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error resolving report", http.StatusInternalServerError)
        return
    }
    // Delete the report after resolving it
    _, err = auth.DB.Exec("DELETE FROM reports WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error deleting resolved report", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Report resolved and deleted successfully"})
}


// Function to allows the admin to reject a report
func RejectReport(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    reportID := r.FormValue("id")
    if reportID == "" {
        http.Error(w, "Report ID is required", http.StatusBadRequest)
        return
    }
    // Update the report status in the database
    _, err := auth.DB.Exec("UPDATE reports SET status = 'rejected' WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error rejecting report", http.StatusInternalServerError)
        return
    }
    // Delete the report after rejecting it
    _, err = auth.DB.Exec("DELETE FROM reports WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error deleting rejected report", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Report rejected and deleted successfully"})
}

// Function to fetches all reports from the database
func GetReports(w http.ResponseWriter, r *http.Request) {
    rows, err := auth.DB.Query(`
        SELECT r.id, r.post_id, r.reason, r.status, p.title, p.content 
        FROM reports r 
        JOIN posts p ON r.post_id = p.id
    `)
    if err != nil {
        http.Error(w, "Error fetching reports", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var reports []map[string]any
    for rows.Next() {
        var id, postID, reason, status, title, content string
        err := rows.Scan(&id, &postID, &reason, &status, &title, &content)
        if err != nil {
            http.Error(w, "Error reading report data", http.StatusInternalServerError)
            return
        }
        report := map[string]any{
            "id":      id,
            "post_id": postID,
            "reason":  reason,
            "status":  status,
            "title":   title,
            "content": content,
        }
        reports = append(reports, report)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(reports)
}

// Function to allows the admin to create a new category
func CreateCategory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    // Retrieve the category name from the request
    categoryName := r.FormValue("name")
    if categoryName == "" {
        http.Error(w, "Category name is required", http.StatusBadRequest)
        return
    }
    // Check if the category already exists
    var existingCategory string
    err := auth.DB.QueryRow("SELECT name FROM categories WHERE name = ?", categoryName).Scan(&existingCategory)
    if err == nil {
        http.Error(w, "Category already exists", http.StatusBadRequest)
        return
    }
    // If a scan error occurs (category not found), continue
    if err != sql.ErrNoRows {
        http.Error(w, "Error checking category existence", http.StatusInternalServerError)
        return
    }
    // Insert the new category into the database
    result, err := auth.DB.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
    if err != nil {
        http.Error(w, "Error creating category", http.StatusInternalServerError)
        return
    }
    // Retrieve the ID of the inserted category
    lastInsertID, err := result.LastInsertId()
    if err != nil {
        http.Error(w, "Error retrieving category ID", http.StatusInternalServerError)
        return
    }
    // Respond with the new category ID
    w.Header().Set("Content-Type", "application/json")

    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Category created successfully",
        "id":      lastInsertID,
    })
}

// Function to allows the admin to delete a category
func DeleteCategory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }
    // Retrieve the category ID to delete
    categoryID := r.FormValue("id")
    if categoryID == "" {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }
    // Delete the category from the database
    _, err := auth.DB.Exec("DELETE FROM categories WHERE id = ?", categoryID)
    if err != nil {
        http.Error(w, "Error deleting category", http.StatusInternalServerError)
        return
    }
    // Respond with a success message
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Category deleted successfully"})
}

// RequestModerator handles a user's request for moderator role
type ModeratorRequest struct {
	UserID string `json:"user_id"`
}

// RequestModerator allows a user to request moderator status
func RequestModerator(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unauthorized method", http.StatusMethodNotAllowed)
		return
	}
	// Retrieve the user ID from the session
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}
	// Record the request in the database
	_, err = auth.DB.Exec("INSERT INTO promotion_requests (user_id, status) VALUES (?, 'pending')", userID)
	if err != nil {
		http.Error(w, "Error submitting request", http.StatusInternalServerError)
		return
	}
	// Respond with a success message
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Promotion request sent successfully"})
}

// Function to allows the admin to view pending moderator requests
func GetModeratorRequests(w http.ResponseWriter, r *http.Request) {
    rows, err := auth.DB.Query("SELECT id, user_id FROM promotion_requests WHERE status = 'pending'")
    if err != nil {
        http.Error(w, "Error fetching requests", http.StatusInternalServerError)
        log.Println("Error fetching requests:", err)
        return
    }
    defer rows.Close()

    var requests []map[string]interface{}
    for rows.Next() {
        var id int
        var userID string

        if err := rows.Scan(&id, &userID); err != nil {
            http.Error(w, "Error reading data", http.StatusInternalServerError)
            log.Println("Error reading rows:", err)
            return
        }
        // Skip invalid user IDs
        if userID == "0" || userID == "" {
            continue
        }
        // Fetch the user's name from the database
        var username string
        err := auth.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
        if err != nil {
            log.Println("Error fetching username:", err)
            username = "Unknown user"
        }

        requests = append(requests, map[string]interface{}{"id": id, "user_id": userID, "username": username})
    }
    if len(requests) == 0 {
        fmt.Fprintln(w, "No pending requests")
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(requests)
}

// Function to promotes a user to moderator
func ApproveModerator(w http.ResponseWriter, r *http.Request) {
    requestID := r.URL.Query().Get("request_id")
    userID := r.URL.Query().Get("user_id")

    if requestID == "" || userID == "" {
        log.Println("Error: Missing request_id or user_id")
        http.Error(w, "Missing request_id or user_id", http.StatusBadRequest)
        return
    }
    // Start a transaction to ensure data integrity
    tx, err := auth.DB.Begin()
    if err != nil {
        log.Println("Error starting transaction:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Update the user's role to 'moderator'
    _, err = tx.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
    if err != nil {
        log.Println("Error updating user role:", err)
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Update the request status to 'approved'
    _, err = tx.Exec("UPDATE promotion_requests SET status = 'approved' WHERE id = ?", requestID)
    if err != nil {
        log.Println("Error updating request status:", err)
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // Commit the transaction
    if err := tx.Commit(); err != nil {
        log.Println("Error committing transaction:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "User has been promoted to moderator")
}

// Function to rejects a moderator promotion request
func RejectModerator(w http.ResponseWriter, r *http.Request) {
    requestID := r.URL.Query().Get("request_id")
    if requestID == "" {
        http.Error(w, "Missing request ID", http.StatusBadRequest)
        return
    }
    // Update the request status to 'rejected'
    _, err := auth.DB.Exec("UPDATE promotion_requests SET status = 'rejected' WHERE id = ?", requestID)
    if err != nil {
        http.Error(w, "Error rejecting request", http.StatusInternalServerError)
        return
    }
    fmt.Fprintln(w, "Request rejected successfully")
}

// Function to allows the admin to manually promote or demote a user
func UpdateUserRole(w http.ResponseWriter, r *http.Request) {
    userID := r.URL.Query().Get("user_id")
    newRole := r.URL.Query().Get("role")

    _, err := auth.DB.Exec("UPDATE users SET role = ? WHERE id = ?", newRole, userID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    fmt.Fprintf(w, "User role updated to %s", newRole)
}

// RemoveModeratorRole - Admin removes moderator role from a user
func RemoveModeratorRole(w http.ResponseWriter, r *http.Request) {

    // Verify if the request method is POST, if not return an error
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    userID := r.FormValue("user_id")
    if userID == "" {
        http.Error(w, "User ID is required", http.StatusBadRequest)
        return
    }
    // Update the user's role to 'user' (remove moderator privileges)
    _, err := auth.DB.Exec("UPDATE users SET role = 'user' WHERE id = ?", userID)
    if err != nil {
        // If an error occurs during the database query, return an internal server error
        http.Error(w, "Error removing moderator role", http.StatusInternalServerError)
        return
    }
    // Respond with a success message in JSON format
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Moderator role removed successfully"})
}

// Function to views the list of all moderators
func GetModerators(w http.ResponseWriter, r *http.Request) {

    // Query the database to get all users who have the 'moderator' role
    rows, err := auth.DB.Query("SELECT id, username FROM users WHERE role = 'moderator'")
    if err != nil {
        http.Error(w, "Error retrieving moderators", http.StatusInternalServerError)
        log.Println("Error retrieving moderators:", err)
        return
    }
    defer rows.Close()

    var moderators []map[string]interface{}
    for rows.Next() {
        var id, username string
        // Scan the result of the query into variables
        if err := rows.Scan(&id, &username); err != nil {
            http.Error(w, "Error reading data", http.StatusInternalServerError)
            log.Println("Error reading database rows:", err)
            return
        }
        // Add the moderator's ID and username to the list of moderators
        moderators = append(moderators, map[string]interface{}{"id": id, "username": username})
    }
    // If no moderators are found, inform the user
    if len(moderators) == 0 {
        fmt.Fprintln(w, "No moderators found")
        return
    }
    // Respond with the list of moderators in JSON format
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(moderators)
}