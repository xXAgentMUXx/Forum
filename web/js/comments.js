document.addEventListener("DOMContentLoaded", function() {
});
function showCommentForm(postID) {
    document.getElementById(`comment-form-${postID}`).style.display = "block";
}
function postComment(postID) {
    let content = document.getElementById(`comment-text-${postID}`).value;
    fetch("/comment/create", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `post_id=${postID}&content=${encodeURIComponent(content)}`
    }).then(() => fetchComments(postID));
}
function fetchComments(postID) {
    fetch(`/comments?post_id=${postID}`)
        .then(response => response.json())
        .then(comments => {
            let commentContainer = document.getElementById(`comments-${postID}`);
            commentContainer.innerHTML = "";
            comments.forEach(comment => {
                let commentID = comment.ID || comment.id; 
                fetchLikeDislikeCount(commentID, "comment", (likeCount, dislikeCount) => {
                    likeCount = likeCount || 0;
                    dislikeCount = dislikeCount || 0;

                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                        <button onclick="likeComment('${commentID}', 'like')">👍 <span id="like-count-${commentID}">${likeCount}</span></button>
                        <button onclick="likeComment('${commentID}', 'dislike')">👎 <span id="dislike-count-${commentID}">${dislikeCount}</span></button>
                        <button onclick="deleteComment('${commentID}')">🗑️ Supprimer</button>
                    `;
                    commentContainer.appendChild(commentElement);
                });
            });
        })
        .catch(error => console.error("Erreur lors du chargement des commentaires :", error));
}


function likeComment(commentID, type) {
    fetch("/like/comment", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${commentID}&type=${type}`
    }).then(() => {
        fetchLikeDislikeCount(commentID, "comment", (likeCount, dislikeCount) => {
            document.getElementById(`like-count-${commentID}`).innerText = likeCount;
            document.getElementById(`dislike-count-${commentID}`).innerText = dislikeCount;
        });
    });
}

function deleteComment(commentID) {
    fetch("/comment/delete", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${commentID}`
    }).then(() => fetchPosts());
}
