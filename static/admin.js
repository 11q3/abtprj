document.addEventListener("DOMContentLoaded", async () => {
    const indicator = document.getElementById("working-indicator");
    const button    = document.getElementById("work-toggle-btn");
    let working     = false;

    // Fetch today's work status
    try {
        const res  = await fetch("/admin/get-work-status");
        const data = await res.json();
        working    = data.working;
        updateUI();
    } catch (e) {
        console.error("Failed to fetch work status:", e);
        indicator.textContent = "ERR";
        button.disabled = true;
    }

    // Intercept all "Mark as Done" forms
    document.querySelectorAll(".complete-form").forEach(form => {
        form.addEventListener("submit", e => {
            if (!working) {
                e.preventDefault();
                alert("Cannot complete a task without an active work session");
            }
        });
    });

    // Work toggle button (start/stop session)
    button.addEventListener("click", async () => {
        const url = working
            ? "/admin/end-work-session"
            : "/admin/start-work-session";

        try {
            const res = await fetch(url, { method: "POST" });
            const text = await res.text();
            if (!res.ok) {
                return alert(text || "Unknown server error");
            }

            // flip state and update UI
            working = !working;
            updateUI();
        } catch (err) {
            console.error("Network error toggling work session:", err);
            alert("Network error, please try again");
        }
    });

    function updateUI() {
        indicator.textContent = working ? "YES" : "NO";
        button.textContent    = working ? "Stop Working" : "Start Working";
    }
});
