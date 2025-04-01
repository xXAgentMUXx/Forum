document.addEventListener("DOMContentLoaded", function () {
    checkSessionAndRedirectToAdmin();
    const postsContainer = document.getElementById("posts");

    // Fonction pour vérifier la session et rediriger si nécessaire
    function checkSessionAndRedirectToAdmin() {
        fetch("/check-session")
            .then(response => {
                if (response.status === 401) {
                    window.location.href = "/login"; // Rediriger vers la connexion si la session est invalide
                    return;
                }
                return response.json(); // Convertir la réponse en JSON
            })
            .then(data => {
                if (!data) return;
    
                console.log("Utilisateur:", data.userID, "| Rôle:", data.role);
    
                // Redirection conditionnelle en fonction du rôle
                if (window.location.pathname === "/admin" && data.role !== "admin") {
                    console.warn("❌ Accès interdit: Vous devez être admin !");
                    window.location.href = "/forbidden"; // Page d'accès interdit
                } else {
                    fetchPosts(); // Charger les posts si l'utilisateur est autorisé
                    fetchComments();
                }
            })
            .catch(error => {
                console.error("Erreur lors de la vérification de la session:", error);
                window.location.href = "/login"; // Rediriger en cas d'erreur
            });
    }

    // Fonction pour récupérer les posts
    async function fetchPosts() {
        try {
            const response = await fetch("/posts");
            if (!response.ok) throw new Error("Erreur lors de la récupération des posts");

            const posts = await response.json();
            if (!Array.isArray(posts)) throw new Error("Données invalides reçues du serveur.");

            displayPosts(posts);
        } catch (error) {
            console.error("Erreur:", error);
            postsContainer.innerHTML = "<p>Impossible de charger les posts.</p>";
        }
    }

    // Fonction pour afficher les posts
    function displayPosts(posts) {
        postsContainer.innerHTML = ""; // Nettoyage avant affichage

        posts.forEach(post => {
            const title = post.Title || "Titre inconnu";
            const content = post.Content || "Aucun contenu disponible.";
            const author = post.Author || "Anonyme";
            const date = post.CreatedAt ? new Date(post.CreatedAt).toLocaleDateString() : "Date inconnue";

            const imageHtml = post.ImagePath && post.ImagePath.trim() !== "" 
            ? `<img src="/${post.ImagePath}" alt="Image du post" style="max-width: 300px; display: block; margin: 0 auto; margin-bottom: 10px;">`
            : "";

            const postElement = document.createElement("div");
            postElement.className = "post";
            postElement.innerHTML = `
                <h2>Post :</h2>
                <h3>${title}</h3>
                <p>${content}</p>
                ${imageHtml}
                <small>Posté par ${author} le ${date}</small>
                <div class="post-buttons">
                    <button class="delete-btn" data-id="${post.ID}">🗑️ Supprimer</button>
                </div>
                <h4>Comments :</h4>
                <div id="comments-${post.ID}" class="comments-container">
                </div>
            `;

            postsContainer.appendChild(postElement);

            // Appeler la fonction pour charger les commentaires du post
            fetchComments(post.ID);

            // Ajouter les gestionnaires d'événements aux boutons
            const deleteButtons = document.querySelectorAll(".delete-btn");

            deleteButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const postID = this.getAttribute("data-id");
                    deletePost(postID);
                });
            });
        });
    }

    // Fonction pour récupérer et afficher les commentaires d'un post
    function fetchComments(postID) {
        fetch(`/comments?post_id=${postID}`) // Récupérer les commentaires depuis le serveur
            .then(response => response.json())
            .then(comments => {
                let commentContainer = document.getElementById(`comments-${postID}`);
                commentContainer.innerHTML = ""; // Effacer les anciens commentaires

                comments.forEach(comment => {
                    let commentID = comment.ID || comment.id;

                    // Créer un nouvel élément de commentaire
                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                        <button class="delete-comment-btn" data-id="${commentID}">🗑️ Supprimer</button>
                    `;

                    // Ajouter le commentaire au conteneur
                    commentContainer.appendChild(commentElement);
                });

                // Ajouter les gestionnaires d'événements pour supprimer les commentaires
                const deleteCommentButtons = document.querySelectorAll(".delete-comment-btn");
                deleteCommentButtons.forEach(button => {
                    button.addEventListener("click", function() {
                        const commentID = this.getAttribute("data-id");
                        deleteComment(commentID);
                    });
                });
            })
            .catch(error => console.error("Erreur lors du chargement des commentaires :", error));
    }

    // Fonction pour supprimer un post
    async function deletePost(postID) {
        if (!confirm("Voulez-vous vraiment supprimer ce post ?")) return;

        try {
            const response = await fetch("/post/delete_admin", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${postID}`
            });

            if (response.ok) {
                alert("Post supprimé !");
                fetchPosts();  // Recharger les posts après suppression
            } else {
                alert("Erreur lors de la suppression !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression du post:", error);
            alert("Une erreur s'est produite.");
        }
    }

    // Fonction pour supprimer un commentaire
    async function deleteComment(commentID) {
        if (!confirm("Voulez-vous vraiment supprimer ce commentaire ?")) return;

        try {
            const response = await fetch("/comments/delete_admin", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${commentID}`
            });

            if (response.ok) {
                alert("Commentaire supprimé !");
                fetchPosts();  // Recharger les posts après suppression
            } else {
                alert("Erreur lors de la suppression du commentaire !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression du commentaire:", error);
            alert("Une erreur s'est produite.");
        }
    }
    // Récupérer et afficher les posts dès que le DOM est chargé
    fetchPosts();
});

document.addEventListener("DOMContentLoaded", function () {
    const reportsContainer = document.getElementById("reports");


async function fetchReports() {
    try {
        const response = await fetch("/report");
        if (!response.ok) throw new Error("Erreur lors de la récupération des rapports");

        const reports = await response.json();
        if (!Array.isArray(reports)) throw new Error("Données invalides reçues du serveur.");

        displayReports(reports);
    } catch (error) {
        console.error("Erreur:", error);
        reportsContainer.innerHTML = "<p>Impossible de charger les rapports.</p>";
    }
}

// Fonction pour afficher les rapports
function displayReports(reports) {
    reportsContainer.innerHTML = ""; // Nettoyage avant affichage

    reports.forEach(report => {
        const postID = report.post_id || "Inconnu";
        const reason = report.reason || "Aucune raison";
        const status = report.status || "En attente";

        const reportElement = document.createElement("div");
        reportElement.className = "report";
        reportElement.innerHTML = `
            <h3>Rapport sur le post ${postID}</h3>
            <p>Raison : ${reason}</p>
            <p>Status : ${status}</p>
            <div class="report-buttons">
                <button class="resolve-btn" data-id="${report.id}">Résoudre</button>
                <button class="reject-btn" data-id="${report.id}">Rejeter</button>
            </div>
        `;

        reportsContainer.appendChild(reportElement);

        // Ajouter les gestionnaires d'événements pour résoudre ou rejeter un rapport
        const resolveButtons = document.querySelectorAll(".resolve-btn");
        const rejectButtons = document.querySelectorAll(".reject-btn");

        resolveButtons.forEach(button => {
            button.addEventListener("click", function() {
                const reportID = this.getAttribute("data-id");
                resolveReport(reportID);
            });
        });

        rejectButtons.forEach(button => {
            button.addEventListener("click", function() {
                const reportID = this.getAttribute("data-id");
                rejectReport(reportID);
            });
        });
    });
}

// Fonction pour résoudre un rapport
async function resolveReport(reportID) {
    try {
        const response = await fetch(`/report/resolve`, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `id=${reportID}`
        });

        if (response.ok) {
            alert("Rapport résolu !");
            fetchReports();  // Recharger les rapports après résolution
        } else {
            alert("Erreur lors de la résolution du rapport !");
        }
    } catch (error) {
        console.error("Erreur lors de la résolution du rapport:", error);
        alert("Une erreur s'est produite.");
    }
}

// Fonction pour rejeter un rapport
async function rejectReport(reportID) {
    try {
        const response = await fetch(`/report/reject`, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `id=${reportID}`
        });

        if (response.ok) {
            alert("Rapport rejeté !");
            fetchReports();  // Recharger les rapports après rejet
        } else {
            alert("Erreur lors du rejet du rapport !");
        }
    } catch (error) {
        console.error("Erreur lors du rejet du rapport:", error);
        alert("Une erreur s'est produite.");
    }
}

// Récupérer et afficher les rapports dès que le DOM est chargé
fetchReports();
});

document.addEventListener("DOMContentLoaded", function () {
    const categoryList = document.getElementById("category-list");
    const createCategoryBtn = document.getElementById("create-category-btn");
    const categoryNameInput = document.getElementById("category-name");

    // Fonction pour récupérer les catégories depuis le serveur
    async function fetchCategories() {
        try {
            const response = await fetch("/categories");
            if (!response.ok) throw new Error("Erreur lors de la récupération des catégories");

            const categories = await response.json();
            if (!Array.isArray(categories)) throw new Error("Données invalides reçues du serveur.");

            displayCategories(categories);
        } catch (error) {
            console.error("Erreur:", error);
            categoryList.innerHTML = "<p>Impossible de charger les catégories.</p>";
        }
    }

    // Fonction pour afficher les catégories
    function displayCategories(categories) {
        categoryList.innerHTML = ""; // Nettoyage avant affichage

        categories.forEach(category => {
            const categoryElement = document.createElement("div");
            categoryElement.className = "category";
            categoryElement.innerHTML = `
                <p>${category.name}</p>
                <button class="delete-category-btn" data-id="${category.id}">🗑️ Supprimer</button>
            `;

            categoryList.appendChild(categoryElement);

            // Ajouter un gestionnaire d'événement pour supprimer une catégorie
            const deleteCategoryButtons = categoryElement.querySelectorAll(".delete-category-btn");
            deleteCategoryButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const categoryID = this.getAttribute("data-id");
                    deleteCategory(categoryID);
                });
            });
        });
    }

    // Fonction pour créer une catégorie
    async function createCategory() {
    const categoryName = categoryNameInput.value.trim();

    if (!categoryName) {
        alert("Le nom de la catégorie ne peut pas être vide");
        return;
    }

    try {
        const response = await fetch("/categories/create", {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `name=${categoryName}`
        });

        if (response.ok) {
            const data = await response.json();
            alert(`Catégorie créée avec succès ! ID de la catégorie : ${data.id}`);
            fetchCategories();  // Recharger les catégories après création
        } else {
            alert("Erreur lors de la création de la catégorie !");
        }
    } catch (error) {
        console.error("Erreur lors de la création de la catégorie:", error);
        alert("Une erreur s'est produite.");
    }
}

    // Fonction pour supprimer une catégorie
    async function deleteCategory(categoryID) {
        if (!confirm("Voulez-vous vraiment supprimer cette catégorie ?")) return;

        try {
            const response = await fetch("/categories/delete", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${categoryID}`
            });

            if (response.ok) {
                alert("Catégorie supprimée !");
                fetchCategories();  // Recharger les catégories après suppression
            } else {
                alert("Erreur lors de la suppression de la catégorie !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression de la catégorie:", error);
            alert("Une erreur s'est produite.");
        }
    }

    // Ajouter un gestionnaire d'événement pour créer une catégorie
    createCategoryBtn.addEventListener("click", createCategory);

    // Récupérer et afficher les catégories dès que le DOM est chargé
    fetchCategories();
});

document.addEventListener("DOMContentLoaded", function () {
    const modRequestList = document.getElementById("mod-request-list");

    // Fonction pour récupérer les demandes de modérateurs
    async function fetchModRequests() {
        try {
            const response = await fetch("/moderator-requests");
            if (!response.ok) {
                const errorText = await response.text();
                throw new Error(`Erreur ${response.status}: ${errorText}`);
            }
    
            const modRequests = await response.json();
            if (!Array.isArray(modRequests)) throw new Error("Données invalides reçues du serveur.");
    
            displayModRequests(modRequests);
        } catch (error) {
            console.error("Erreur:", error);
            modRequestList.innerHTML = `<p>Impossible de charger les demandes.</p>`;
        }
    }

    // Fonction pour afficher les demandes de modérateurs
    function displayModRequests(modRequests) {
        modRequestList.innerHTML = ""; // Nettoyage avant affichage
    
        modRequests.forEach(request => {
            const userID = request.user_id || "Inconnu"; // Assurez-vous d'avoir un userID
            const username = request.username || "Nom d'utilisateur inconnu";  // Affichage du nom d'utilisateur
            const status = request.status || "En attente";
    
            const requestElement = document.createElement("div");
            requestElement.className = "mod-request";
            requestElement.innerHTML = `
                <h3>Demande de modération de ${username}</h3>
                <p>ID utilisateur: ${userID}</p>
                <p>Status: ${status}</p>
                <div class="mod-request-buttons">
                    <button class="accept-btn" data-id="${request.id}" data-user-id="${userID}">Accepter</button>
                    <button class="reject-btn" data-id="${request.id}" data-user-id="${userID}">Rejeter</button>
                </div>
            `;
            modRequestList.appendChild(requestElement);
    
            // Ajouter les gestionnaires d'événements pour accepter ou rejeter la demande
            const acceptButtons = requestElement.querySelectorAll(".accept-btn");
            const rejectButtons = requestElement.querySelectorAll(".reject-btn");
    
            acceptButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const requestID = this.getAttribute("data-id");
                    const userID = this.getAttribute("data-user-id");  // Récupérer l'userID ici
                    handleModRequest(requestID, "accept", userID);  // Passer userID à la fonction
                });
            });
    
            rejectButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const requestID = this.getAttribute("data-id");
                    const userID = this.getAttribute("data-user-id");  // Récupérer l'userID ici
                    handleModRequest(requestID, "reject", userID);  // Passer userID à la fonction
                });
            });
        });
    }
    

    // Fonction pour accepter ou rejeter une demande de modérateur
    async function handleModRequest(requestID, action, userID) {
        const url = action === "accept" ? `/approve-moderator?request_id=${requestID}&user_id=${userID}` : `/reject-moderator?request_id=${requestID}&user_id=${userID}`;
        
        try {
            const response = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `request_id=${requestID}&user_id=${userID}`  // Ajouter user_id dans le corps de la requête
            });
    
            if (response.ok) {
                alert(`Demande ${action}ée !`);
                fetchModRequests();  // Recharger les demandes après traitement
            } else {
                alert("Erreur lors du traitement de la demande !");
            }
        } catch (error) {
            console.error("Erreur:", error);
            alert("Une erreur s'est produite.");
        }
    }
    

    // Récupérer et afficher les demandes de modérateurs dès que le DOM est chargé
    fetchModRequests();
});

document.addEventListener("DOMContentLoaded", function () {
    const roleUpdateForm = document.getElementById("role-update-form");

    // Fonction pour mettre à jour le rôle d'un utilisateur
    roleUpdateForm.addEventListener("submit", async function(event) {
        event.preventDefault(); // Empêcher l'envoi du formulaire

        const userId = document.getElementById("user-id").value;
        const newRole = document.getElementById("new-role").value;

        if (!userId || !newRole) {
            alert("Veuillez remplir tous les champs.");
            return;
        }

        try {
            const response = await fetch("/update-role", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `user_id=${userId}&role=${newRole}`
            });

            if (response.ok) {
                alert("Rôle mis à jour !");
            } else {
                alert("Erreur lors de la mise à jour du rôle !");
            }
        } catch (error) {
            console.error("Erreur:", error);
            alert("Une erreur s'est produite.");
        }
    });
});

async function loadModerators() {
    const response = await fetch("/get-moderators");
    if (response.ok) {
        const moderators = await response.json();
        const moderatorList = document.getElementById("moderator-list");

        moderatorList.innerHTML = ""; // Clear existing list

        moderators.forEach(moderator => {
            const moderatorItem = document.createElement("div");
            moderatorItem.classList.add("moderator-item");
            moderatorItem.innerHTML = `
                <span>${moderator.username}</span>
                <button onclick="removeModeratorRole('${moderator.id}')">Retirer modérateur</button>
            `;
            moderatorList.appendChild(moderatorItem);
        });
    } else {
        alert("Erreur lors du chargement des modérateurs");
    }
}

async function removeModeratorRole(userID) {
    if (!confirm("Voulez-vous vraiment retirer le rôle de modérateur ?")) return;

    try {
        const response = await fetch("/remove-moderator-role", {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `user_id=${userID}`
        });

        if (response.ok) {
            alert("Rôle de modérateur retiré !");
            loadModerators();  // Reload the moderators list
        } else {
            alert("Erreur lors du retrait du rôle de modérateur !");
        }
    } catch (error) {
        console.error("Erreur lors du retrait du rôle de modérateur:", error);
        alert("Une erreur s'est produite.");
    }
}

// Load the list of moderators when the page is ready
document.addEventListener("DOMContentLoaded", loadModerators);