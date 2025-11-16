// Countdown timer
const targetDate = new Date("2025-12-31T00:00:00Z");

function updateCountdown() {
    const now = new Date();
    const difference = targetDate - now;
    
    if (difference < 0) {
        document.getElementById("countdown").innerHTML = "<p style='color: white; font-size: 1.5rem;'>We're Live!</p>";
        return;
    }
    
    const days = Math.floor(difference / (1000 * 60 * 60 * 24));
    const hours = Math.floor((difference / (1000 * 60 * 60)) % 24);
    const minutes = Math.floor((difference / (1000 * 60)) % 60);
    const seconds = Math.floor((difference / 1000) % 60);
    
    document.getElementById("days").textContent = days.toString().padStart(2, "0");
    document.getElementById("hours").textContent = hours.toString().padStart(2, "0");
    document.getElementById("minutes").textContent = minutes.toString().padStart(2, "0");
    document.getElementById("seconds").textContent = seconds.toString().padStart(2, "0");
}

setInterval(updateCountdown, 1000);
updateCountdown(); // Run immediately
