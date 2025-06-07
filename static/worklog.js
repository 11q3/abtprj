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
    
})

