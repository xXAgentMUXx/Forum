let lastRequestTime = 0;
const requestLimit = 3000; 

document.addEventListener("DOMContentLoaded", () => {
    const now = Date.now();

    if (now - lastRequestTime < requestLimit) {
        alert("Vous chargez la page trop rapidement. Veuillez patienter.");
        return;
    }

    lastRequestTime = now;
});