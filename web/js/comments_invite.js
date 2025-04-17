// Event listener that triggers when the DOM content is fully loaded
document.addEventListener("DOMContentLoaded", function() {
});

// Function to display the comment form
function showCommentForm(postID) {
    document.getElementById(`comment-form-${postID}`).style.display = "block";
}

// Function to post a comment
function postComment(postID) {
    let content = document.getElementById(`comment-text-${postID}`).value;
    fetch("/comment/create", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `post_id=${postID}&content=${encodeURIComponent(content)}`
    }).then(() => fetchComments(postID));
}

// Function to fetch and display comments
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

                    // Create a new div element for the comment
                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                        👍 <span id="like-count-${commentID}">${likeCount}</span>
                        👎 <span id="dislike-count-${commentID}">${dislikeCount}</span>
                    `;
                    // Append the new comment to the container
                    commentContainer.appendChild(commentElement);
                });
            });
        })
        .catch(error => console.error("Erreur lors du chargement des commentaires :", error));
}