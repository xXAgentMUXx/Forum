// Fetch the user's activity data once the page is ready
document.addEventListener("DOMContentLoaded", function() {
    checkSessionAndFetchActivity();
});

// Function to prevent going to the link without connecting to the forum
function checkSessionAndFetchActivity() {
    fetch("/check-session")
        .then(response => {
            if (response.status === 401) { 
                window.location.href = "/login"; 
            } else {
                fetchActivity(); 
            }
        })
        .catch(error => {
            console.error("Erreur lors de la vérification de la session:", error);
            window.location.href = "/login"; 
        });
}

// Asynchronous function to fetch the user's activity
async function fetchActivity() {
    try {
        let response = await fetch("/user/activity");
        let data = await response.json();

        // Display user posts
        let postContainer = document.getElementById("my-posts");
        if (data.posts && data.posts.length > 0) {
            postContainer.innerHTML = ""; 
            data.posts.forEach(post => {
                let div = document.createElement("div");
                div.classList.add("post");
                div.innerHTML = `<h3>${post.title}</h3><p>${post.content}</p><small>${new Date(post.created_at).toLocaleString()}</small>`;
                postContainer.appendChild(div);
            });
        } else {
            postContainer.innerHTML = "<p>Aucun post trouvé.</p>";
        }

        // Display likes and dislikes (posts + commentaires)
        let likeContainer = document.getElementById("my-likes");
        likeContainer.innerHTML = ""; 

        if ((data.likes && data.likes.length > 0) || (data.comment_likes && data.comment_likes.length > 0)) {
            // Likes/Dislikes des posts
            if (data.likes && data.likes.length > 0) {
                data.likes.forEach(like => {
                    let div = document.createElement("div");
                    div.classList.add(like.type === "like" ? "like" : "dislike");
                    div.innerHTML = `<p>Vous avez ${like.type === "like" ? "aimé" : "disliké"} le post : <strong>${like.title}</strong></p>`;
                    likeContainer.appendChild(div);
                });
            }

            // Likes/Dislikes des commentaires
            if (data.comment_likes && data.comment_likes.length > 0) {
                data.comment_likes.forEach(like => {
                    let div = document.createElement("div");
                    div.classList.add(like.type === "like" ? "like" : "dislike");
                    div.innerHTML = `<p>Vous avez ${like.type === "like" ? "aimé" : "disliké"} sur le post <strong>${like.post_title}</strong> du commentaire : "${like.comment}"</p>`;
                    likeContainer.appendChild(div);
                });
            }
        } else {
            likeContainer.innerHTML = "<p>Aucun like ou dislike trouvé.</p>";
        }

        // Display user comments
        let commentContainer = document.getElementById("my-comments");
        if (data.comments && data.comments.length > 0) {
            commentContainer.innerHTML = ""; 
            data.comments.forEach(comment => {
                let div = document.createElement("div");
                div.classList.add("comment");
                div.innerHTML = `<p>Commenté sur : <strong>${comment.title}</strong></p><p>"${comment.comment}"</p><small>${new Date(comment.created_at).toLocaleString()}</small>`;
                commentContainer.appendChild(div);
            });
        } else {
            commentContainer.innerHTML = "<p>Aucun commentaire trouvé.</p>";
        }

    } catch (error) {
        console.error("Erreur lors du chargement de l'activité :", error);
    }
}