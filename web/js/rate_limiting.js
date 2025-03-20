// Event listener triggered when the DOM content is fully loaded
document.addEventListener("DOMContentLoaded", function () {
    const params = new URLSearchParams(window.location.search);
    const errorMessage = document.getElementById("error-message");

    // Check if there is an "error" parameter in the URL
    if (params.has("error")) {
        errorMessage.style.display = "block";
          
        // Switch case to display specific error messages based on the error type
        switch (params.get("error")) {
            case "toomany":  // Too many attempts error
                errorMessage.textContent = "Trop de tentatives. Veuillez patienter avant de réessayer.";
                break;
            case "invalid": // Invalid credentials error
                errorMessage.textContent = "Identifiants incorrects. Veuillez réessayer.";
                break;
        }
    }
});