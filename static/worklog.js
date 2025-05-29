document.addEventListener("DOMContentLoaded", async () => {
    // --- 1) Format DONE AT timestamps ---
    document.querySelectorAll('.done-at').forEach(span => {
        const raw = span.textContent.trim();
        if (!raw || raw === '-') return;
        const d = new Date(raw);
        span.textContent = d.toLocaleString('en-US', {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: 'numeric',
            minute: 'numeric',
            second: 'numeric',
            hour12: false
        });
    });

    // --- 2) Date picker reload on change ---
    const picker = document.getElementById("date");
    const params = new URLSearchParams(window.location.search);
    const today = new Date().toISOString().split("T")[0];
    picker.value = params.get("date") || today;
    picker.addEventListener("change", function () {
        window.location.href = `/worklog/?date=${this.value}`;
    });

    // --- 3) Working status ---
    const indicator = document.getElementById("working-indicator");
    const button = document.getElementById("work-toggle-btn");

    let working = false;

    try {
        const res = await fetch("/admin/get-work-status");
        const data = await res.json();

        working = data.working;
        updateUI();
    } catch (e) {
        console.error("Failed to fetch status", e);
    }

    function updateUI() {
        indicator.textContent = working ? "YES" : "NO";
        button.textContent = working ? "Stop Working" : "Start Working";
    }

    button.addEventListener("click", async () => {
        const url = working ? "/admin/end-work-session" : "/admin/start-work-session";
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
