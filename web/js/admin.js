document.addEventListener("DOMContentLoaded", function () {
    const postsContainer = document.getElementById("posts");

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
                ? `<img src="/${post.ImagePath}" alt="Image du post" style="max-width: 300px;">`
                : "";

            const postElement = document.createElement("div");
            postElement.className = "post";
            postElement.innerHTML = `
                <h3>${title}</h3>
                <p>${content}</p>
                <small>Posté par ${author} le ${date}</small>
                ${imageHtml}
                <div class="post-buttons">
                    <button class="delete-btn" data-id="${post.ID}">🗑️ Supprimer</button>
                </div>
                <div id="comments-${post.ID}" class="comments-container">
                    <!-- Les commentaires seront chargés ici -->
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
