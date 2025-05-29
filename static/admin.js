document.addEventListener("DOMContentLoaded", async () => {
    const indicator = document.getElementById("working-indicator");
    const button = document.getElementById("work-toggle-btn");

    let working = false;

    try {
        const res = await fetch("/admin/get-work-status");
        const data = await res.json();

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

    function updateUI() {
        indicator.textContent = working ? "YES" : "NO";
        button.textContent = working ? "Stop Working" : "Start Working";
    }

    button.addEventListener("click", async () => {
        const url = working ? "/admin/stop-work-session" : "/admin/start-work-session";
        try {
            const res = await fetch(url, { method: "POST" });
            if (res.ok) {
                working = !working;
                updateUI();
            } else {
                console.error("Session toggle failed", await res.text());
            }
        } catch (e) {
            console.error("Toggle error", e);
        }
    });
});