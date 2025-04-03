document.addEventListener("DOMContentLoaded", function () {
    checkSessionAndRedirectToModerator();
    const postsContainer = document.getElementById("posts");

    function checkSessionAndRedirectToModerator() {
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
                if (window.location.pathname === "/moderator" && data.role !== "moderator") {
                    console.warn("❌ Accès interdit: Vous devez être moderateur !");
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
                <small style="display: block; margin-top: 10px;">Posté par ${author} le ${date}</small>
                <div class="post-buttons">
                    <button class="delete-btn" data-id="${post.ID}">🗑️ Supprimer</button>
                    <button class="report-btn" data-id="${post.ID}">⚠️ Signaler</button>
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
            const reportButtons = document.querySelectorAll(".report-btn");

            deleteButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const postID = this.getAttribute("data-id");
                    deletePost(postID);
                });
            });

            reportButtons.forEach(button => {
                button.addEventListener("click", function() {
                    const postID = this.getAttribute("data-id");
                    reportPost(postID);
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
                    `;

                    // Ajouter le commentaire au conteneur
                    commentContainer.appendChild(commentElement);
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

    // Fonction pour signaler un post
    async function reportPost(postID) {
        const reason = prompt("Pourquoi signalez-vous ce post ?");
        if (!reason) return;
        
        try {
            const response = await fetch("/report/post", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${postID}&moderator_id=1&reason=${encodeURIComponent(reason)}` // Remplacez 1 par l'ID du modérateur réel
            });
    
            if (response.ok) {
                alert("Post signalé à l'administration !");
            } else {
                alert("Erreur lors du signalement !");
            }
        } catch (error) {
            console.error("Erreur lors du signalement du post:", error);
            alert("Une erreur s'est produite.");
        }
    }
    // Récupérer et afficher les posts dès que le DOM est chargé
    fetchPosts();
});