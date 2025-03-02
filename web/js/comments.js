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
                fetchLikeDislikeCount(comment.ID, "comment", (likeCount, dislikeCount) => {
                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                        <button onclick="likeComment('${comment.id}', 'like')">👍 <span id="like-count-${comment.id}">${likeCount}</span></button>
                        <button onclick="likeComment('${comment.id}', 'dislike')">👎 <span id="dislike-count-${comment.id}">${dislikeCount}</span></button>
                        <button onclick="deleteComment('${comment.id}')">🗑️ Supprimer</button>
                    `;
                    commentContainer.appendChild(commentElement);
                });
            });
        });
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
