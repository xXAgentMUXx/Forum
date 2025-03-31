// Event listener that triggers when the DOM content is fully loaded
document.addEventListener("DOMContentLoaded", function() {
    fetchPosts();
    checkSessionAndFetchPosts();
});

// Function to check if the user session is valid
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

// Function to fetch posts based on selected filter
function fetchPosts(filter = "all", categoryID = "") {
    let url = "/posts";
    if (filter === "category" && categoryID) {
        url += `?filter=category&category_id=${categoryID}`;
    } else if (filter === "my_posts") {
        url += "?filter=my_posts";
    } else if (filter === "liked") {
        url += "?filter=liked";
    }
    loadCategories();  // Load categories
    fetch(url) // Fetch posts
    .then(response => response.json())
    .then(posts => {
        let postContainer = document.getElementById("posts");
        postContainer.innerHTML = "";
        posts.forEach(post => {
            fetchLikeDislikeCount(post.ID, "post", function(likeCount, dislikeCount) {
                let postElement = document.createElement("div");
                postElement.classList.add("post");

                let imageHtml = "";
                // If post has an image, display it
                if (post.ImagePath && post.ImagePath.trim() !== "") {
                    imageHtml = `<img src="/${post.ImagePath}" alt="Post Image" style="max-width: 300px;">`;
                }
                // Create post HTML structure
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
                fetchComments(post.ID); // Fetch and display comments
            });
        });
    })
    .catch(error => console.error("Erreur lors du chargement des posts :", error));
}

// Function to request moderator status for a user
function requestModerator(userID) {
    fetch("/request-moderator", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ user_id: userID })
    })
    .then(response => response.json())
    .then(data => console.log("Réponse du serveur:", data))
    .catch(error => console.error("Erreur:", error));
}

// Function to load categories for the filter and posts
function loadCategories() {
    fetch("/categories")
        .then(response => response.json())
        .then(categories => {
            let filterSelect = document.getElementById("post-category-dropdown");
            let postFormSelect = document.getElementById("post-category");

            // Check if category dropdowns are found
            if (!filterSelect || !postFormSelect) {
                console.error("❌ Erreur : Un des menus de sélection des catégories est introuvable !");
                return;
            }
            let optionsHTML = `<option value="">Sélectionner une catégorie</option>`;
            categories.forEach(category => {
                optionsHTML += `<option value="${category.id}">${category.name}</option>`;
            });

            filterSelect.innerHTML = optionsHTML;   // Populate category dropdown for filter
            postFormSelect.innerHTML = optionsHTML; // Populate category dropdown for posts
        })
        .catch(error => console.error("❌ Erreur lors du chargement des catégories :", error));
}

// Function to fetch like and dislike counts
function fetchLikeDislikeCount(contentID, contentType, callback) {
    if (typeof callback !== "function") {
        console.error("Erreur : callback non défini pour fetchLikeDislikeCount");
        return;
    }
    // Fetch likes and dislikes count
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

// Function to apply selected filter on posts
function applyFilter() {
    let filter = document.getElementById("filter").value;
    let categorySelect = document.getElementById("post-category-dropdown");

    if (!categorySelect) {
        console.error("❌ Erreur : Le menu déroulant de catégorie est introuvable !");
        return;
    }
    let categoryContainer = categorySelect.parentElement; 

     // Display category filter if "category" filter is selected, and mask others
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

// Function to cancel post creation
function cancelPostCreation() {
    document.getElementById("post-form").style.display = "none";
    document.getElementById("post-title").value = "";
    document.getElementById("post-content").value = "";
    document.getElementById("post-category").selectedIndex = 0;
    document.getElementById("post-image").value = "";
    document.getElementById("image-preview").style.display = "none";
}

// Function to delete a post
function deletePost(postID) {
    fetch("/post/delete", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${postID}`
    }).then(() => fetchPosts());
}

// Function to create a new post
async function createPost() {
    const title = document.getElementById("post-title").value;
    const content = document.getElementById("post-content").value;
    const categorySelect = document.getElementById("post-category");
    const imageInput = document.getElementById("post-image");

    let selectedCategories = Array.from(categorySelect.selectedOptions).map(option => option.value);
    
    // Validate form fields
    if (!title || !content || selectedCategories.length === 0) {
        alert("Veuillez remplir tous les champs.");
        return;
    }
    const formData = new FormData();
    formData.append("title", title);
    formData.append("content", content);
    formData.append("categories", selectedCategories.join(",")); 

    if (imageInput.files.length > 0) {
        formData.append("image", imageInput.files[0]);
    }
    try {
        const response = await fetch("/post/create", {
            method: "POST",
            body: formData
        });
        // Reload posts after successful creation
        if (response.ok) {
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

// Function to like or dislike a post
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

// Function to toggle the display of the post creation form
function showPostForm() {
    let form = document.getElementById("post-form");
    if (form.style.display === "none") {
        form.style.display = "block";
    } else {
        form.style.display = "none";
    }
}

function previewImage(event) {
    const file = event.target.files[0];
    const preview = document.getElementById("preview-img");
    const previewContainer = document.getElementById("image-preview");

    if (file) {
        const reader = new FileReader();
        reader.onload = function(e) {
            preview.src = e.target.result;
            previewContainer.style.display = "block";
        };
        reader.readAsDataURL(file);
    } else {
        previewContainer.style.display = "none";
    }
}

// Function to preview the selected image for post
function removeImage() {
    const fileInput = document.getElementById("post-image");
    const previewContainer = document.getElementById("image-preview");
    // Clear image input
    fileInput.value = ""; 
    previewContainer.style.display = "none"; 
}
