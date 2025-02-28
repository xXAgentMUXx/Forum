document.addEventListener("DOMContentLoaded", function () {
    // Inscription
    document.getElementById("registerForm")?.addEventListener("submit", async function (event) {
        event.preventDefault();
        const email = document.getElementById("email").value;
        const username = document.getElementById("username").value;
        const password = document.getElementById("password").value;

        const response = await fetch("/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ Email: email, Username: username, Password: password }),
        });

        const result = await response.json();
        document.getElementById("registerMessage").textContent = result.message;
    });

    // Connexion
    document.getElementById("loginForm")?.addEventListener("submit", async function (event) {
        event.preventDefault();
        const email = document.getElementById("loginEmail").value;
        const password = document.getElementById("loginPassword").value;

        const response = await fetch("/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ Email: email, Password: password }),
        });

        const result = await response.json();
        document.getElementById("loginMessage").textContent = result.message;
    });
});