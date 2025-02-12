document.getElementById("notify-button").addEventListener("click", async () => {
  const emailInput = document.getElementById("email-input");
  const email = emailInput.value.trim();

  if (email) {
    try {
      const response = await fetch("/subscribe", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email }),
      });

      // Check if response is OK, then parse JSON
      const result = await response.json();
      console.log("Server response:", result);
      alert(result.message || result.error);
    } catch (error) {
      console.error("Error during subscribe fetch:", error);
      alert("There was an error during subscription. Please try again later.");
    }
    emailInput.value = ""; // Clear input field
  } else {
    alert("Please enter a valid email.");
  }
});

window.addEventListener('scroll', function() {
    var footerHeight = document.querySelector('footer').offsetHeight;
    if (window.scrollY + window.innerHeight >= document.body.offsetHeight - footerHeight) {
            document.querySelector('footer').style.display = 'block';
        } else {
            document.querySelector('footer').style.display = 'none';
        }
});
