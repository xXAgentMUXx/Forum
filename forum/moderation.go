package forum

import (
	"Forum/auth"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)


func ServeModerator(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/moderator.html")
}

func ServeAdmin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "web/html/admin.html")
}

func DeletePostByAdmin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Récupérer l'ID du post à supprimer
	postID := r.FormValue("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	// Vérifier si le post existe dans la base de données
	var postOwner string
	err := auth.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postOwner)
	if err == sql.ErrNoRows {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Error retrieving post", http.StatusInternalServerError)
		return
	}

	// Supprimer le post de la base de données
	_, err = auth.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		http.Error(w, "Error deleting post", http.StatusInternalServerError)
		return
	}

	// Répondre avec un message de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
}

func DeleteCommentAdmin(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Récupérer l'ID du commentaire
    commentID := r.FormValue("id")
    if commentID == "" {
        http.Error(w, "Comment ID is required", http.StatusBadRequest)
        return
    }

    // Vérifier si le commentaire existe dans la base de données
    var commentOwner string
    err := auth.DB.QueryRow("SELECT user_id FROM comments WHERE id = ?", commentID).Scan(&commentOwner)
    if err == sql.ErrNoRows {
        http.Error(w, "Comment not found", http.StatusNotFound)
        return
    } else if err != nil {
        http.Error(w, "Error retrieving comment", http.StatusInternalServerError)
        return
    }

    // Supprimer le commentaire de la base de données
    _, err = auth.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
    if err != nil {
        http.Error(w, "Error deleting comment", http.StatusInternalServerError)
        return
    }

    fmt.Fprintf(w, "Comment deleted successfully!")
}

func ReportPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier si la connexion à la DB est active
	if auth.DB == nil {
		log.Println("Database connection is nil")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Récupération des données du formulaire
	postID := r.FormValue("id")
	reason := r.FormValue("reason")

	if postID == "" || reason == "" {
		log.Println("Missing parameters: postID or reason")
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Debugging : Afficher les valeurs récupérées
	log.Println("Post ID:", postID, "Reason:", reason)

	// Insérer dans la base de données avec un log
	query := "INSERT INTO reports (post_id, reason, status) VALUES (?, ?, 'pending')"
	log.Println("Executing SQL Query:", query)

	_, err := auth.DB.Exec(query, postID, reason)
	if err != nil {
		log.Println("Database error:", err)
		http.Error(w, "Error creating report", http.StatusInternalServerError)
		return
	}

	// Répondre avec un message de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Report submitted successfully"})
}


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

    // Mettre à jour le statut du rapport dans la base de données
    _, err := auth.DB.Exec("UPDATE reports SET status = 'resolved' WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error resolving report", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Report resolved successfully"})
}

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

    // Mettre à jour le statut du rapport dans la base de données
    _, err := auth.DB.Exec("UPDATE reports SET status = 'rejected' WHERE id = ?", reportID)
    if err != nil {
        http.Error(w, "Error rejecting report", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Report rejected successfully"})
}
func GetReports(w http.ResponseWriter, r *http.Request) {
    rows, err := auth.DB.Query("SELECT id, post_id, reason, status FROM reports")
    if err != nil {
        http.Error(w, "Error fetching reports", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var reports []map[string]interface{}
    for rows.Next() {
        var id, postID, reason, status string
        err := rows.Scan(&id, &postID, &reason, &status)
        if err != nil {
            http.Error(w, "Error reading report data", http.StatusInternalServerError)
            return
        }
        report := map[string]interface{}{
            "id":     id,
            "post_id": postID,
            "reason": reason,
            "status": status,
        }
        reports = append(reports, report)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(reports)
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Récupérer le nom de la catégorie depuis la requête
    categoryName := r.FormValue("name")
    if categoryName == "" {
        http.Error(w, "Category name is required", http.StatusBadRequest)
        return
    }

    // Vérification si la catégorie existe déjà
    var existingCategory string
    err := auth.DB.QueryRow("SELECT name FROM categories WHERE name = ?", categoryName).Scan(&existingCategory)
    if err == nil {
        http.Error(w, "Category already exists", http.StatusBadRequest)
        return
    }

    // Insérer la nouvelle catégorie dans la base de données
    _, err = auth.DB.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
    if err != nil {
        http.Error(w, "Error creating category", http.StatusInternalServerError)
        return
    }

    // Répondre avec un message de succès
    w.Header().Set("Content-Type", "application/json")
    log.Println("Nom de la catégorie reçu :", categoryName)

    json.NewEncoder(w).Encode(map[string]string{"message": "Category created successfully"})
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // Récupérer l'ID de la catégorie à supprimer
    categoryID := r.FormValue("id")
    if categoryID == "" {
        http.Error(w, "Category ID is required", http.StatusBadRequest)
        return
    }

    // Supprimer la catégorie de la base de données
    _, err := auth.DB.Exec("DELETE FROM categories WHERE id = ?", categoryID)
    if err != nil {
        http.Error(w, "Error deleting category", http.StatusInternalServerError)
        return
    }

    // Répondre avec un message de succès
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Category deleted successfully"})
}