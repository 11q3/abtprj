document.addEventListener("DOMContentLoaded", async () => {
    const indicator = document.getElementById("working-indicator");
    const button = document.getElementById("work-toggle-btn");

    try {
        const res = await fetch("/work-status");
        const data = await res.json();
        const working = data.working;

        if (working) {
            indicator.textContent = "YES";
            button.textContent = "Stop Working";
        } else {
            indicator.textContent = "NO";
            button.textContent = "Start Working";
        }
    } catch (e) {
        console.error("Failed to fetch status", e);
    }
});