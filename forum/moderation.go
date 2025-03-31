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

    var reports []map[string]any
    for rows.Next() {
        var id, postID, reason, status string
        err := rows.Scan(&id, &postID, &reason, &status)
        if err != nil {
            http.Error(w, "Error reading report data", http.StatusInternalServerError)
            return
        }
        report := map[string]any{
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

    // Si une erreur de scan se produit (catégorie non trouvée), on continue
    if err != sql.ErrNoRows {
        http.Error(w, "Error checking category existence", http.StatusInternalServerError)
        return
    }

    // Insérer la nouvelle catégorie dans la base de données
    result, err := auth.DB.Exec("INSERT INTO categories (name) VALUES (?)", categoryName)
    if err != nil {
        http.Error(w, "Error creating category", http.StatusInternalServerError)
        return
    }

    // Récupérer l'ID de la catégorie insérée
    lastInsertID, err := result.LastInsertId()
    if err != nil {
        http.Error(w, "Error retrieving category ID", http.StatusInternalServerError)
        return
    }

    // Répondre avec un message de succès et l'ID de la nouvelle catégorie
    w.Header().Set("Content-Type", "application/json")
    log.Println("Nom de la catégorie reçu :", categoryName)

    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Category created successfully",
        "id":      lastInsertID,
    })
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

// RequestModerator - User requests moderator role
type ModeratorRequest struct {
	UserID string `json:"user_id"`
}


// Fonction pour gérer la demande de modération
func RequestModerator(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	// Récupérer l'ID utilisateur depuis la session
	userID, err := auth.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}
	// Log de l'ID utilisateur
	log.Printf("Request Moderator - user_id: %s", userID)

	// Enregistrer la demande en base de données
	_, err = auth.DB.Exec("INSERT INTO promotion_requests (user_id, status) VALUES (?, 'pending')", userID)
	if err != nil {
		http.Error(w, "Erreur lors de l'envoi de la demande", http.StatusInternalServerError)
		return
	}
	// Répondre avec un message de succès
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Demande de promotion envoyée avec succès"})

}


// GetModeratorRequests - Admin views pending requests
func GetModeratorRequests(w http.ResponseWriter, r *http.Request) {
    rows, err := auth.DB.Query("SELECT id, user_id FROM promotion_requests WHERE status = 'pending'")
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des demandes", http.StatusInternalServerError)
        log.Println("Erreur lors de la récupération des demandes:", err)
        return
    }
    defer rows.Close()

    var requests []map[string]interface{}
    for rows.Next() {
        var id int
        var userID string // userID est maintenant une chaîne (UUID ou autre)

        if err := rows.Scan(&id, &userID); err != nil {
            http.Error(w, "Erreur lors de la lecture des données", http.StatusInternalServerError)
            log.Println("Erreur lors de la lecture des lignes de la base de données:", err)
            return
        }

        // Ignorez les demandes avec un user_id invalide
        if userID == "0" || userID == "" {
            continue // Passez cette ligne si user_id est invalide
        }

        // Chercher le nom de l'utilisateur dans la base de données (exemple avec la table "users")
        var username string
        err := auth.DB.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
        if err != nil {
            log.Println("Erreur lors de la récupération du nom d'utilisateur:", err)
            username = "Nom d'utilisateur inconnu"
        }

        // Si userID est valide, ajoutez-le à la liste des demandes
        requests = append(requests, map[string]interface{}{"id": id, "user_id": userID, "username": username})
    }

    if len(requests) == 0 {
        fmt.Fprintln(w, "Aucune demande en attente")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(requests)
}

// ApproveModerator - Admin approves request and promotes user
func ApproveModerator(w http.ResponseWriter, r *http.Request) {
    requestID := r.URL.Query().Get("request_id")
    userID := r.URL.Query().Get("user_id")

    // Log des paramètres reçus
    log.Printf("Received requestID: %s, userID: %s", requestID, userID)

    if requestID == "" || userID == "" {
        log.Println("Error: Missing request_id or user_id")
        http.Error(w, "ID de demande ou ID utilisateur manquant", http.StatusBadRequest)
        return
    }

    // Démarrer une transaction pour assurer l'intégrité des données
    tx, err := auth.DB.Begin()
    if err != nil {
        log.Println("Error starting transaction:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mettre à jour le rôle de l'utilisateur
    _, err = tx.Exec("UPDATE users SET role = 'moderator' WHERE id = ?", userID)
    if err != nil {
        log.Println("Error updating user role:", err)
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Mettre à jour l'état de la demande
    _, err = tx.Exec("UPDATE promotion_requests SET status = 'approved' WHERE id = ?", requestID)
    if err != nil {
        log.Println("Error updating request status:", err)
        tx.Rollback()
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Valider la transaction
    if err := tx.Commit(); err != nil {
        log.Println("Error committing transaction:", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    fmt.Fprintln(w, "L'utilisateur a été promu au rôle de modérateur")
}

func RejectModerator(w http.ResponseWriter, r *http.Request) {
    requestID := r.URL.Query().Get("request_id")
    if requestID == "" {
        http.Error(w, "ID de demande manquant", http.StatusBadRequest)
        return
    }
    // Mettre à jour l'état de la demande à 'rejected'
    _, err := auth.DB.Exec("UPDATE promotion_requests SET status = 'rejected' WHERE id = ?", requestID)
    if err != nil {
        http.Error(w, "Erreur lors du rejet de la demande", http.StatusInternalServerError)
        return
    }
    fmt.Fprintln(w, "Demande rejetée avec succès")
}

// UpdateUserRole - Admin promotes or demotes a user manually
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
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    userID := r.FormValue("user_id")
    if userID == "" {
        http.Error(w, "User  ID is required", http.StatusBadRequest)
        return
    }

    // Mettre à jour le rôle de l'utilisateur
    _, err := auth.DB.Exec("UPDATE users SET role = 'user' WHERE id = ?", userID)
    if err != nil {
        http.Error(w, "Error removing moderator role", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Moderator role removed successfully"})
}

func GetModerators(w http.ResponseWriter, r *http.Request) {
    rows, err := auth.DB.Query("SELECT id, username FROM users WHERE role = 'moderator'")
    if err != nil {
        http.Error(w, "Erreur lors de la récupération des modérateurs", http.StatusInternalServerError)
        log.Println("Erreur lors de la récupération des modérateurs:", err)
        return
    }
    defer rows.Close()

    var moderators []map[string]interface{}
    for rows.Next() {
        var id, username string
        if err := rows.Scan(&id, &username); err != nil {
            http.Error(w, "Erreur lors de la lecture des données", http.StatusInternalServerError)
            log.Println("Erreur lors de la lecture des lignes de la base de données:", err)
            return
        }

        moderators = append(moderators, map[string]interface{}{"id": id, "username": username})
    }

    if len(moderators) == 0 {
        fmt.Fprintln(w, "Aucun modérateur trouvé")
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(moderators)
}