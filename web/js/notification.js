document.addEventListener("DOMContentLoaded", function () {
    fetchNotifications();
    setInterval(fetchNotifications, 10000);
});

function toggleNotificationBox() {
    console.log("toggleNotificationBox() appelé !");
    let notifBox = document.getElementById("notification-box");

    console.log("État actuel:", notifBox.classList);

    if (notifBox.classList.contains("hidden")) {
        notifBox.classList.remove("hidden");
        notifBox.style.display = "block";
        markNotificationsAsSeen();
        
        // Appeler fetchNewComments pour chaque post dans les notifications
        const postIDs = Array.from(document.querySelectorAll('.notification-item a')).map(link => link.href.split('/').pop());
        postIDs.forEach(postID => fetchNewComments(postID));
    } else {
        notifBox.classList.add("hidden");
        notifBox.style.display = "none";
    }

    console.log("Nouvel état:", notifBox.classList);
}

function fetchNotifications() {
    fetch("/notifications")
        .then(response => response.json())
        .then(notifications => {
            console.log("Notifications reçues :", notifications);

            let notifIcon = document.getElementById("notification-icon");
            let notifDropdown = document.getElementById("notification-dropdown");

            notifDropdown.innerHTML = '';  // Effacer les notifications existantes avant d'ajouter les nouvelles

            if (notifications.length === 0) {
                notifDropdown.innerHTML += "<p>Aucune notification</p>";
                return;
            }

            notifications.forEach(notif => {
                let notifElement = document.createElement("div");
                notifElement.classList.add("notification-item");
                notifElement.id = `notif-${notif.id}`; // Ajout de l'ID de la notification pour référence

                let deleteButton = document.createElement("button");
                deleteButton.innerText = "Supprimer";
                deleteButton.classList.add("delete-notif");
                deleteButton.onclick = () => deleteNotification(notif.id, notifElement); // Passer l'élément DOM à supprimer

                if (notif.action === "comment") {
                    let shortContent = notif.content.length > 50 ? notif.content.substring(0, 50) + "..." : notif.content;
                    notifElement.innerHTML = `
                        <p><strong>${notif.username}</strong> a commenté votre post</p>
                        <p>"${shortContent}"</p>
                        <small>${new Date(notif.created_at).toLocaleString()}</small>
                    `;
                } else if (notif.action === "like") {
                    notifElement.innerHTML = `
                        <p><strong>${notif.username}</strong> a ${notif.content} votre post/commentaire</p>
                        <small>${new Date(notif.created_at).toLocaleString()}</small>
                    `;
                }

                notifElement.appendChild(deleteButton);
                notifDropdown.appendChild(notifElement);
            });
        })
        .catch(error => console.error("Erreur lors de la récupération des notifications :", error));
}

function markNotificationsAsSeen() {
    fetch("/notifications/mark-seen", { method: "POST" })
        .then(() => fetchNotifications())
        .catch(error => console.error("Erreur lors de la mise à jour des notifications :", error));
}

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

function deleteNotification(notifID, notifElement) {
    fetch("/notifications/delete", {
        method: "POST",
        headers: { "Content-Type": "application/x-www-form-urlencoded" },
        body: `id=${notifID}`
    })
    .then(response => {
        if (response.ok) {
            // Supprimer l'élément du DOM après une suppression réussie
            notifElement.remove();
        } else {
            console.error("Erreur lors de la suppression de la notification");
        }
    })
    .catch(error => console.error("Erreur lors de la suppression :", error));
}
