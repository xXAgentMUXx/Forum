document.addEventListener("DOMContentLoaded", function () {
    checkSessionAndRedirectToAdmin();
    const postsContainer = document.getElementById("posts");

    // Function to redirect to the /forum for security, if session is invalid
    function checkSessionAndRedirectToAdmin() {
        fetch("/check-session")
            .then(response => {
                if (response.status === 401) {
                    window.location.href = "/login"; 
                    return;
                }
                return response.json(); 
            })
            .then(data => {
                if (!data) return;
    
                console.log("Utilisateur:", data.userID, "| Rôle:", data.role);
    
                // Redirect in function of the role of the user
                if (window.location.pathname === "/admin" && data.role !== "admin") {
                    console.warn("❌ Accès interdit: Vous devez être admin !");
                    window.location.href = "/forbidden"; 
                } else {
                    fetchPosts(); 
                    fetchComments();
                }
            })
            .catch(error => {
                console.error("Erreur lors de la vérification de la session:", error);
                window.location.href = "/login"; 
            });
    }

    // Function to fetch the posts
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

    // Function to display the posts
    function displayPosts(posts) {
        postsContainer.innerHTML = ""; 

        // Display the templates
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

            // Call function to load commentary
            fetchComments(post.ID);

            // Add delete bouton
            const deleteButtons = document.querySelectorAll(".delete-btn");

            deleteButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const postID = this.getAttribute("data-id");
                    deletePost(postID);
                });
            });
        });
    }

    // Function to fetch comments
    function fetchComments(postID) {
        fetch(`/comments?post_id=${postID}`) 
            .then(response => response.json())
            .then(comments => {
                let commentContainer = document.getElementById(`comments-${postID}`);
                commentContainer.innerHTML = ""; 

                // Display the comments
                comments.forEach(comment => {
                    let commentID = comment.ID || comment.id;


                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                        <button class="delete-comment-btn" data-id="${commentID}">🗑️ Supprimer</button>
                    `;

                    // Add to the container
                    commentContainer.appendChild(commentElement);
                });

                // Add event for the commentary
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

    // Function to delete a post
    async function deletePost(postID) {
        if (!confirm("Voulez-vous vraiment supprimer ce post ?")) return;

        try {
            const response = await fetch("/post/delete_admin", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${postID}`
            });
            // reload posts after suppresion
            if (response.ok) {
                alert("Post supprimé !");
                fetchPosts();  
            } else {
                alert("Erreur lors de la suppression !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression du post:", error);
            alert("Une erreur s'est produite.");
        }
    }

    // Function to delete commentary
    async function deleteComment(commentID) {
        if (!confirm("Voulez-vous vraiment supprimer ce commentaire ?")) return;

        try {
            const response = await fetch("/comments/delete_admin", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${commentID}`
            });
            // Reload comments after suprresion
            if (response.ok) {
                alert("Commentaire supprimé !");
                fetchPosts(); 
            } else {
                alert("Erreur lors de la suppression du commentaire !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression du commentaire:", error);
            alert("Une erreur s'est produite.");
        }
    }
    // Retrieves posts if DOM is charged
    fetchPosts();
});

document.addEventListener("DOMContentLoaded", function () {
    const reportsContainer = document.getElementById("reports");

// Function to fetch reports
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

// Function to display the reports
function displayReports(reports) {
    reportsContainer.innerHTML = ""; 

    // Display the templates
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

        // Add event for the reports like resolve or reject it
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

// Function to resolve report
async function resolveReport(reportID) {
    try {
        const response = await fetch(`/report/resolve`, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `id=${reportID}`
        });

        if (response.ok) {
            alert("Rapport résolu !");
            fetchReports();  // Reload report after resolution
        } else {
            alert("Erreur lors de la résolution du rapport !");
        }
    } catch (error) {
        console.error("Erreur lors de la résolution du rapport:", error);
        alert("Une erreur s'est produite.");
    }
}

// Function to reject report
async function rejectReport(reportID) {
    try {
        const response = await fetch(`/report/reject`, {
            method: "POST",
            headers: { "Content-Type": "application/x-www-form-urlencoded" },
            body: `id=${reportID}`
        });

        if (response.ok) {
            alert("Rapport rejeté !");
            fetchReports();  // Reload report after reject
        } else {
            alert("Erreur lors du rejet du rapport !");
        }
    } catch (error) {
        console.error("Erreur lors du rejet du rapport:", error);
        alert("Une erreur s'est produite.");
    }
}

// Retrieves the report after DOM is loaded
fetchReports();
});

document.addEventListener("DOMContentLoaded", function () {
    const categoryList = document.getElementById("category-list");
    const createCategoryBtn = document.getElementById("create-category-btn");
    const categoryNameInput = document.getElementById("category-name");

    // Function to retrieves categories
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

    // Function to display categories
    function displayCategories(categories) {
        categoryList.innerHTML = "";

        // Display the templates
        categories.forEach(category => {
            const categoryElement = document.createElement("div");
            categoryElement.className = "category";
            categoryElement.innerHTML = `
                <p>${category.name}</p>
                <button class="delete-category-btn" data-id="${category.id}">🗑️ Supprimer</button>
            `;

            categoryList.appendChild(categoryElement);

            // Add event for the categories
            const deleteCategoryButtons = categoryElement.querySelectorAll(".delete-category-btn");
            deleteCategoryButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const categoryID = this.getAttribute("data-id");
                    deleteCategory(categoryID);
                });
            });
        });
    }

    // Function to cretae the categories
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
        // Reload categorie after creation
        if (response.ok) {
            const data = await response.json();
            alert(`Catégorie créée avec succès ! ID de la catégorie : ${data.id}`);
            fetchCategories();  
        } else {
            alert("Erreur lors de la création de la catégorie !");
        }
    } catch (error) {
        console.error("Erreur lors de la création de la catégorie:", error);
        alert("Une erreur s'est produite.");
    }
}

    // Function to delete a categorie
    async function deleteCategory(categoryID) {
        if (!confirm("Voulez-vous vraiment supprimer cette catégorie ?")) return;

        try {
            const response = await fetch("/categories/delete", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${categoryID}`
            });
            //Reload categories after suprresion
            if (response.ok) {
                alert("Catégorie supprimée !");
                fetchCategories();  
            } else {
                alert("Erreur lors de la suppression de la catégorie !");
            }
        } catch (error) {
            console.error("Erreur lors de la suppression de la catégorie:", error);
            alert("Une erreur s'est produite.");
        }
    }

    // Add event for the categories
    createCategoryBtn.addEventListener("click", createCategory);

    // Retrieves categories after DOM is loaded
    fetchCategories();
});

document.addEventListener("DOMContentLoaded", function () {
    const modRequestList = document.getElementById("mod-request-list");

    // Fonction to fetch moderator request
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

    // Fonction to display moderator request
    function displayModRequests(modRequests) {
        modRequestList.innerHTML = ""; // Nettoyage avant affichage
        
        // Display the templates
        modRequests.forEach(request => {
            const userID = request.user_id || "Inconnu"; 
            const username = request.username || "Nom d'utilisateur inconnu";  
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
    
            // Add event for the request
            const acceptButtons = requestElement.querySelectorAll(".accept-btn");
            const rejectButtons = requestElement.querySelectorAll(".reject-btn");
    
            acceptButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const requestID = this.getAttribute("data-id");
                    const userID = this.getAttribute("data-user-id");  
                    handleModRequest(requestID, "accept", userID);  
                });
            });
    
            rejectButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const requestID = this.getAttribute("data-id");
                    const userID = this.getAttribute("data-user-id");  
                    handleModRequest(requestID, "reject", userID);  
                });
            });
        });
    }
    
    // Function to accepte or reject moderator request
    async function handleModRequest(requestID, action, userID) {
        const url = action === "accept" ? `/approve-moderator?request_id=${requestID}&user_id=${userID}` : `/reject-moderator?request_id=${requestID}&user_id=${userID}`;
        
        try {
            const response = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `request_id=${requestID}&user_id=${userID}`  
            });
            // reload request after treatements
            if (response.ok) {
                alert(`Demande ${action}ée !`);
                fetchModRequests();  
            } else {
                alert("Erreur lors du traitement de la demande !");
            }
        } catch (error) {
            console.error("Erreur:", error);
            alert("Une erreur s'est produite.");
        }
    }
    

    //Retrives request after DOM is loaded
    fetchModRequests();
});

document.addEventListener("DOMContentLoaded", function () {
    const roleUpdateForm = document.getElementById("role-update-form");

    // Function to update the new role
    roleUpdateForm.addEventListener("submit", async function(event) {
        event.preventDefault(); 

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

// Function to load the moderator
async function loadModerators() {
    const response = await fetch("/get-moderators");
    if (response.ok) {
        const moderators = await response.json();
        const moderatorList = document.getElementById("moderator-list");

        moderatorList.innerHTML = ""; 

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

// function to delete moderator role
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