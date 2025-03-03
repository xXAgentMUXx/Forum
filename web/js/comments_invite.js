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
                       👍 <span id="like-count-${comment.id}">${likeCount}</span></button>
                       👎 <span id="dislike-count-${comment.id}">${dislikeCount}</span></button>
                        
                    `;
                    commentContainer.appendChild(commentElement);
                });
            });
        });
}


