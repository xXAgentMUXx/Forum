document.addEventListener("DOMContentLoaded", function() {
    fetchPosts();
});

function fetchPosts(filter = "all", categoryID = "") {
    let url = "/posts";
    if (filter === "category" && categoryID) {
        url += `?filter=category&category_id=${categoryID}`;
    } else if (filter === "my_posts") {
        url += "?filter=my_posts";
    } else if (filter === "liked") {
        url += "?filter=liked";
    }

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
                        imageHtml = `<img src="/${post.ImagePath}" alt="Post Image"  class="post-image">`;
                    }
                    postElement.innerHTML = `
                        <h2>${post.Title}</h2>
                        <p>${post.Content}</p>
                        ${imageHtml}
                        <div class="like-dislike-buttons"> 
                        👍 <span id="like-count-${post.ID}">${likeCount}</span></button>
                        👎 <span id="dislike-count-${post.ID}">${dislikeCount}</span></button>
                        </div>
                        <div id="comments-${post.ID}"></div>
                        <div id="comment-form-${post.ID}" style="display:none;">
                        </div>
                    `;
                    postContainer.appendChild(postElement);
                    fetchComments(post.ID);
                });
            });
        });
}

function fetchLikeDislikeCount(contentID, contentType, callback) {
    fetch(`/likes?id=${contentID}&type=${contentType}`)
        .then(response => response.json())
        .then(data => {
            console.log(`Likes pour ${contentType} ${contentID} :`, data); 
            callback(data.likes || 0, data.dislikes || 0);
        })
        .catch(error => {
            console.error("Erreur lors de la récupération des likes/dislikes :", error);
            callback(0, 0);
        });
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
