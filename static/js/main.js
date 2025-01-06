document.getElementById("notify-button").addEventListener("click", async () => {
    const emailInput = document.getElementById("email-input");
    const email = emailInput.value;

    if (email) {
        const response = await fetch("/subscribe", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email }),
        });

        const result = await response.json();
        alert(result.message || result.error);
        emailInput.value = ""; // Clear input field
    } else {
        alert("Please enter a valid email.");
    }
});
