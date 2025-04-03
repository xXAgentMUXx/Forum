// Fetch notifications every 10 seconds
document.addEventListener("DOMContentLoaded", function () {
    fetchNotifications();
    setInterval(fetchNotifications, 10000);

    // check if you connect to forum
    if (window.location.pathname === "/forum") {
        closeNotificationBox();
    }
});

// function to close the box when you are in the forum
function closeNotificationBox() {
    let notifBox = document.getElementById("notification-box");
    notifBox.classList.add("hidden");
    notifBox.style.display = "none";
}

// Toggles the visibility of the notification box
function toggleNotificationBox() {
    console.log("toggleNotificationBox() appelé !");
    let notifBox = document.getElementById("notification-box");

    console.log("État actuel:", notifBox.classList);

    // If the notification box is hidden, make it visible
    if (notifBox.classList.contains("hidden")) {
        notifBox.classList.remove("hidden");
        notifBox.style.display = "block";
        markNotificationsAsSeen(); // Mark notifications as seen when the box is shown
        
        // Fetch new comments for each post linked in the notifications
        const postIDs = Array.from(document.querySelectorAll('.notification-item a')).map(link => link.href.split('/').pop());
        postIDs.forEach(postID => fetchNewComments(postID));
    } else {
        // Otherwise, hide the notification box
        notifBox.classList.add("hidden");
        notifBox.style.display = "none";
    }

    console.log("Nouvel état:", notifBox.classList);
}

// Fetches the list of notifications
function fetchNotifications() {
    fetch("/notifications")
        .then(response => response.json())
        .then(notifications => {
            console.log("Notifications reçues :", notifications);

            let notifIcon = document.getElementById("notification-icon");
            let notifDropdown = document.getElementById("notification-dropdown");

            notifDropdown.innerHTML = '';  

            if (notifications.length === 0) {
                notifDropdown.innerHTML += "<p>Aucune notification</p>";
                return;
            }
            // Display each notification
            notifications.forEach(notif => {
                let notifElement = document.createElement("div");
                notifElement.classList.add("notification-item");
                notifElement.id = `notif-${notif.id}`; 

                let deleteButton = document.createElement("button");
                deleteButton.innerText = "Supprimer";
                deleteButton.classList.add("delete-notif");
                deleteButton.onclick = () => deleteNotification(notif.id, notifElement); 

                // Check if username contain email adress
                let username = notif.username;
                if (
                    username.includes("@gmail.com") ||
                    username.includes("@ynov.com") ||
                    username.includes("@yahoo.com") ||
                    username.includes("@outlook.com") ||
                    username.includes("@hotmail.com") ||
                    username.includes("@icloud.com") ||
                    username.includes("@live.com") ||
                    username.includes("@aol.com") ||
                    username.includes("@yandex.com") ||
                    username.includes("@protonmail.com") ||
                    username.includes("@zoho.com") ||
                    username.includes("@mail.com") ||
                    username.includes("@msn.com") ||
                    username.includes("@ymail.com") ||
                    username.includes("@comcast.net") ||
                    username.includes("@mail.ru") ||
                    username.includes("@gnail.com") ||
                    username.includes("@tutanota.com") ||
                    username.includes("@fastmail.com") ||
                    username.includes("@gmx.com")
                ) {
                    // Replace by this word if there is an email
                    username = "Quelqu'un";
                }
                // Fetch If the notification is for a comment or a like
                if (notif.action === "comment") {
                    let shortContent = notif.content.length > 50 ? notif.content.substring(0, 50) + "..." : notif.content;
                    notifElement.innerHTML = `
                        <p><strong>${username}</strong> a commenté votre post</p>
                        <p>"${shortContent}"</p>
                        <small>${new Date(notif.created_at).toLocaleString()}</small>
                    `;
                } else if (notif.action === "like") {
                    notifElement.innerHTML = `
                        <p><strong>${username}</strong> a ${notif.content} votre post/commentaire</p>
                        <small>${new Date(notif.created_at).toLocaleString()}</small>
                    `;
                }
                // Add the delete button to the notification element
                notifElement.appendChild(deleteButton);
                notifDropdown.appendChild(notifElement);
            });
        })
        .catch(error => console.error("Erreur lors de la récupération des notifications :", error));
}

// Marks all notifications as seen
function markNotificationsAsSeen() {
    fetch("/notifications/mark-seen", { method: "POST" })
        .then(() => fetchNotifications())
        .catch(error => console.error("Erreur lors de la mise à jour des notifications :", error));
}
// Fetches new comments for a given post
function fetchNewComments(postID) {
    console.log(`Fetching new comments for post ID: ${postID}`);
    fetch(`/comments/new?post_id=${postID}`)
        .then(response => response.json())
        .then(comments => {
            console.log("Nouveaux commentaires :", comments); 
            if (comments && comments.length > 0) {
                let notifDropdown = document.getElementById("notification-dropdown");
                comments.forEach(comment => {
                    let commentElement = document.createElement("div");
                    commentElement.classList.add("notification-item");
                    commentElement.innerHTML = `
                        <p><strong>Nouveau commentaire</strong></p>
                        <p>${comment.content}</p>
                        <small>Posté le ${new Date(comment.created_at).toLocaleString()}</small>
                    `;
                    notifDropdown.appendChild(commentElement);
                });
            }
        })
        .catch(error => console.error("Erreur lors de la récupération des nouveaux commentaires :", error));
}

// Deletes a specific notification
function deleteNotification(notifID, notifElement) {
    fetch("/notifications/delete", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${notifID}`
    })
    // Remove the notification element from the DOM after successful deletion
    .then(response => {
        if (response.ok) {
            notifElement.remove();
        } else {
            console.error("Erreur lors de la suppression de la notification");
        }
    })
    .catch(error => console.error("Erreur lors de la suppression :", error));
}
