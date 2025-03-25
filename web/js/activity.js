// Fetch the user's activity data once the page is ready
document.addEventListener("DOMContentLoaded", function() {
    fetchActivity();
});

// Asynchronous function to fetch the user's activity
async function fetchActivity() {
    try {
         // Send a GET request to the '/user/activity' endpoint
        let response = await fetch("/user/activity");
        let data = await response.json();
        
        // Display the posts in the designated container
        let postContainer = document.getElementById("my-posts");
        data.posts.forEach(post => {
            let div = document.createElement("div");
            div.classList.add("post");
            div.innerHTML = `<h3>${post.title}</h3><p>${post.content}</p><small>${new Date(post.created_at).toLocaleString()}</small>`;
            postContainer.appendChild(div);
        });

        // Display the likes/dislike in the designated container
        let likeContainer = document.getElementById("my-likes");
        data.likes.forEach(like => {
            let div = document.createElement("div");
            div.classList.add(like.type === "like" ? "like" : "dislike");
            div.innerHTML = `<p>Vous avez ${like.type === "like" ? "aimé" : "disliké"} : <strong>${like.title}</strong></p>`;
            likeContainer.appendChild(div);
        });

        // Display the comments in the designated container
        let commentContainer = document.getElementById("my-comments");
        data.comments.forEach(comment => {
            let div = document.createElement("div");
            div.classList.add("comment");
            div.innerHTML = `<p>Commenté sur : <strong>${comment.title}</strong></p><p>"${comment.comment}"</p><small>${new Date(comment.created_at).toLocaleString()}</small>`;
            commentContainer.appendChild(div);
        });

    } catch (error) {
        console.error("Erreur lors du chargement de l'activité :", error);
    }
}
