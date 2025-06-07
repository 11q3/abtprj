// static/admin.js

document.addEventListener("DOMContentLoaded", () => {
    const indicator = document.getElementById("working-indicator");
    const button    = document.getElementById("work-toggle-btn");

    // 0) Initialise from server-rendered span ("YES"/"NO")
    let working = indicator.textContent.trim().toUpperCase() === "YES";

    // 1) Wire the toggle
    button.addEventListener("click", async () => {
        const url = working
            ? "/admin/end-work-session"
            : "/admin/start-work-session";

        try {
            const res = await fetch(url, {
                method:      "POST",
                credentials: "same-origin"
            });
            if (!res.ok) throw new Error(`HTTP ${res.status}`);

            // flip state
            working = !working;

            // 2) update **both** indicator and button
            indicator.textContent = working ? "YES" : "NO";
            button.textContent    = working ? "Stop Working" : "Start Working";

        } catch (err) {
            console.error("Toggle failed:", err);
            alert("Couldn’t toggle session. Check console.");
        }
    });

    // 3) (Optional) Prevent completing tasks when not working
    document.querySelectorAll(".complete-form").forEach(form => {
        form.addEventListener("submit", e => {
            if (!working) {
                e.preventDefault();
                alert("Cannot complete a task without an active work session");
            }
        });
    });
});
