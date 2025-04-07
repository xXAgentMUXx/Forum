document.addEventListener("DOMContentLoaded", function () {
    checkSessionAndRedirectToModerator();
    const postsContainer = document.getElementById("posts");

    function checkSessionAndRedirectToModerator() {
        fetch("/check-session")
            .then(response => {
                if (response.status === 401) {
                    window.location.href = "/"; // Redirect to login if the session is invalid
                    return;
                }
                return response.json(); // Convert the response to JSON
            })
            .then(data => {
                if (!data) return;
    
                console.log("User:", data.userID, "| Role:", data.role);
    
                // Conditional redirection based on user role
                if (window.location.pathname === "/moderator" && data.role !== "moderator") {
                    console.warn("❌ Access denied: You must be a moderator!");
                    window.location.href = "/forbidden"; 
                } else {
                    fetchPosts(); // Load posts if the user is authorized
                    fetchComments();
                }
            })
            .catch(error => {
                console.error("Error during session check:", error);
                window.location.href = "/"; 
            });
    }
    
    // Function to fetch posts
    async function fetchPosts() {
        try {
            const response = await fetch("/posts");
            if (!response.ok) throw new Error("Error fetching posts");

            const posts = await response.json();
            if (!Array.isArray(posts)) throw new Error("Invalid data received from the server.");

            displayPosts(posts);
        } catch (error) {
            console.error("Error:", error);
            postsContainer.innerHTML = "<p>Unable to load posts.</p>";
        }
    }

    // Function to display posts
    function displayPosts(posts) {
        postsContainer.innerHTML = ""; 

        // Display posts
        posts.forEach(post => {
            const title = post.Title || "Unknown title";
            const content = post.Content || "No content available.";
            const author = post.Author || "Anonymous";
            const date = post.CreatedAt ? new Date(post.CreatedAt).toLocaleDateString() : "Unknown date";
    
            const imageHtml = post.ImagePath && post.ImagePath.trim() !== "" 
            ? `<img src="/${post.ImagePath}" alt="Post image" style="max-width: 300px; display: block; margin: 0 auto; margin-bottom: 10px;">`
            : "";

            const postElement = document.createElement("div");
            postElement.className = "post";
            postElement.innerHTML = `
                <h2>Post:</h2>
                <h3>${title}</h3>
                <p>${content}</p>
                ${imageHtml}
                <small style="display: block; margin-top: 10px;">Posted by ${author} on ${date}</small>
                <div class="post-buttons">
                    <button class="delete-btn" data-id="${post.ID}">🗑️ Delete</button>
                    <button class="report-btn" data-id="${post.ID}">⚠️ Report</button>
                </div>
                <h4>Comments:</h4>
                <div id="comments-${post.ID}" class="comments-container">
                </div>
            `;    

            postsContainer.appendChild(postElement);

            // Call function to load comments for the post
            fetchComments(post.ID);

            // Add event listeners to the buttons
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

    // Function to fetch and display comments for a post
    function fetchComments(postID) {
        fetch(`/comments?post_id=${postID}`) // Fetch comments from the server
            .then(response => response.json())
            .then(comments => {
                let commentContainer = document.getElementById(`comments-${postID}`);
                commentContainer.innerHTML = ""; // Clear previous comments

                comments.forEach(comment => {
                    let commentID = comment.ID || comment.id;

                    // Create a new comment element
                    let commentElement = document.createElement("div");
                    commentElement.classList.add("comment");
                    commentElement.innerHTML = `
                        <p>${comment.content}</p>
                    `;

                    // Add the comment to the container
                    commentContainer.appendChild(commentElement);
                });
            })
            .catch(error => console.error("Error loading comments:", error));
    }

    // Function to delete a post
    async function deletePost(postID) {
        if (!confirm("Are you sure you want to delete this post?")) return;

        try {
            const response = await fetch("/post/delete_admin", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${postID}`
            });
            // Reload posts after deletion
            if (response.ok) {
                alert("Post deleted!");
                fetchPosts();  
            } else {
                alert("Error deleting the post!");
            }
        } catch (error) {
            console.error("Error deleting the post:", error);
            alert("An error occurred.");
        }
    }

    // Function to report a post
    async function reportPost(postID) {
        const reason = prompt("Why are you reporting this post?");
        if (!reason) return;
        
        try {
            const response = await fetch("/report/post", {
                method: "POST",
                headers: { "Content-Type": "application/x-www-form-urlencoded" },
                body: `id=${postID}&moderator_id=1&reason=${encodeURIComponent(reason)}` 
            });
    
            if (response.ok) {
                alert("Post reported to the administration!");
            } else {
                alert("Error reporting the post!");
            }
        } catch (error) {
            console.error("Error reporting the post:", error);
            alert("An error occurred.");
        }
    }
    // Fetch and display posts as soon as the DOM is loaded
    fetchPosts();
});