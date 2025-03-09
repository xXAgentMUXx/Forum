document.addEventListener("DOMContentLoaded", function () {
    const params = new URLSearchParams(window.location.search);
    const errorMessage = document.getElementById("error-message");

    if (params.has("error")) {
        errorMessage.style.display = "block";
        switch (params.get("error")) {
            case "toomany":
                errorMessage.textContent = "Trop de tentatives. Veuillez patienter avant de réessayer.";
                break;
            case "invalid":
                errorMessage.textContent = "Identifiants incorrects. Veuillez réessayer.";
                break;
        }
    }
});