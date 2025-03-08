document.addEventListener("DOMContentLoaded", function() {
    fetchPosts();
    checkSessionAndFetchPosts()
});
function checkSessionAndFetchPosts() {
    fetch("/check-session") 
        .then(response => {
            if (response.status === 401) { 
                window.location.href = "/login";
            } 
        })
        .catch(error => {
            console.error("Erreur lors de la vérification de la session:", error);
            window.location.href = "/login"; 
        });
}
function fetchPosts(filter = "all", categoryID = "") {
    let url = "/posts";
    if (filter === "category" && categoryID) {
        url += `?filter=category&category_id=${categoryID}`;
    } else if (filter === "my_posts") {
        url += "?filter=my_posts";
    } else if (filter === "liked") {
        url += "?filter=liked";
    }

    loadCategories();
    fetch(url)
    .then(response => response.json())
    .then(posts => {
        let postContainer = document.getElementById("posts");
        postContainer.innerHTML = "";
        posts.forEach(post => {
            fetchLikeDislikeCount(post.ID, "post", function(likeCount, dislikeCount) {
                let postElement = document.createElement("div");
                postElement.classList.add("post");

                let imageHtml = "";
                if (post.ImagePath && post.ImagePath.trim() !== "") {
                    imageHtml = `<img src="/${post.ImagePath}" alt="Post Image" style="max-width: 300px;">`;
                }
                postElement.innerHTML = `
                     <h2>${post.Title}</h2>
                    <p>${post.Content}</p>
                    ${imageHtml}
                    <div class="post-buttons">
                    <button onclick="likePost('${post.ID}', 'like')">👍 <span id="like-count-${post.ID}">${likeCount}</span></button>
                    <button onclick="likePost('${post.ID}', 'dislike')">👎 <span id="dislike-count-${post.ID}">${dislikeCount}</span></button>
                    <button onclick="showCommentForm('${post.ID}')">Commenter</button>
                    <button onclick="deletePost('${post.ID}')">🗑️ Supprimer</button>
                    </div>
                    <div id="comments-${post.ID}"></div>
                    <div id="comment-form-${post.ID}" style="display:none;">
                    <textarea id="comment-text-${post.ID}" placeholder="Votre commentaire"></textarea>
                    <button onclick="postComment('${post.ID}')">Publier</button>
                    </div>
                `;
                postContainer.appendChild(postElement);
                fetchComments(post.ID);
            });
        });
    })
    .catch(error => console.error("Erreur lors du chargement des posts :", error));
}

function loadCategories() {
    fetch("/categories")
        .then(response => response.json())
        .then(categories => {
            let filterSelect = document.getElementById("post-category-dropdown");
            let postFormSelect = document.getElementById("post-category");

            if (!filterSelect || !postFormSelect) {
                console.error("❌ Erreur : Un des menus de sélection des catégories est introuvable !");
                return;
            }

            let optionsHTML = `<option value="">Sélectionner une catégorie</option>`;
            categories.forEach(category => {
                optionsHTML += `<option value="${category.id}">${category.name}</option>`;
            });

            filterSelect.innerHTML = optionsHTML;   
            postFormSelect.innerHTML = optionsHTML; 
        })
        .catch(error => console.error("❌ Erreur lors du chargement des catégories :", error));
}

function fetchLikeDislikeCount(contentID, contentType, callback) {
    if (typeof callback !== "function") {
        console.error("Erreur : callback non défini pour fetchLikeDislikeCount");
        return;
    }

    fetch(`/likes?id=${contentID}&type=${contentType}`)
        .then(response => response.json())
        .then(data => {
            callback(data.likes || 0, data.dislikes || 0);
        })
        .catch(error => {
            console.error("Erreur lors de la récupération des likes/dislikes :", error);
            callback(0, 0);
        });
}

function applyFilter() {
    let filter = document.getElementById("filter").value;
    let categorySelect = document.getElementById("post-category-dropdown");

    if (!categorySelect) {
        console.error("❌ Erreur : Le menu déroulant de catégorie est introuvable !");
        return;
    }
    let categoryContainer = categorySelect.parentElement; 

    if (filter === "category") {
        categoryContainer.style.display = "inline-block"; 
    } else {
        categoryContainer.style.display = "none"; 
    }
    let categoryID = categorySelect.value;

    if (filter !== "category") {
        categoryID = ""; 
    }
    fetchPosts(filter, categoryID);
}
function deletePost(postID) {
    fetch("/post/delete", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${postID}`
    }).then(() => fetchPosts());
}

async function createPost() {
    const title = document.getElementById("post-title").value;
    const content = document.getElementById("post-content").value;
    const category = document.getElementById("post-category").value;
    const imageInput = document.getElementById("post-image");
    
    if (!title || !content || !category) {
        alert("Veuillez remplir tous les champs.");
        return;
    }
    
    const formData = new FormData();
    formData.append("title", title);
    formData.append("content", content);
    formData.append("category_id", category);
    
    if (imageInput.files.length > 0) {
        formData.append("image", imageInput.files[0]);
    }
    
    try {
        const response = await fetch("/post/create", {
            method: "POST",
            body: formData
        });
        
        if (response.ok) {
            alert("Post créé avec succès !");
            fetchPosts(); 
        } else {
            const errorMessage = await response.text();
            alert("Erreur: " + errorMessage);
        }
    } catch (error) {
        console.error("Erreur lors de la création du post", error);
        alert("Une erreur s'est produite.");
    }
}

function likePost(postID, type) {
    fetch("/like/post", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${postID}&type=${type}`
    }).then(() => {
        fetchLikeDislikeCount(postID, "post", (likeCount, dislikeCount) => {
            document.getElementById(`like-count-${postID}`).innerText = likeCount;
            document.getElementById(`dislike-count-${postID}`).innerText = dislikeCount;
        });
    });
}
function showPostForm() {
    let form = document.getElementById("post-form");
    if (form.style.display === "none") {
        form.style.display = "block";
    } else {
        form.style.display = "none";
    }
}

