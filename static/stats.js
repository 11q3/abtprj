document.addEventListener("DOMContentLoaded", () => {
    const detailBox = document.getElementById("goal-details");

    function escapeHTML(str) {
        const div = document.createElement('div');
        div.textContent = str || '';
        return div.innerHTML;
    }

    document.querySelectorAll(".contrib-graph .day").forEach(day => {
        const goalNodes = day.querySelectorAll('.goal-data');
        if (!goalNodes.length) return;
        day.addEventListener('click', e => {
            e.stopPropagation();
            const goals = Array.from(goalNodes).map(g => ({
                name: g.dataset.name,
                desc: g.dataset.desc,
                status: g.dataset.status,
                done: g.dataset.done,
                due: g.dataset.due
            }));
            showDetails(goals, day);
        });
    });

    document.addEventListener('click', e => {
        if (!detailBox.contains(e.target)) {
            detailBox.style.display = 'none';
        }
    });

    function showDetails(goals, cell) {
        detailBox.innerHTML = goals.map(g => {
            return `<div class="goal-entry"><strong>${escapeHTML(g.name)}</strong>`+
                `<div>${escapeHTML(g.desc)}</div>`+
                `<div>Status: ${escapeHTML(g.status)}</div>`+
                `${g.done ? `<div>Done at: ${escapeHTML(g.done)}</div>` : `<div>Due at: ${escapeHTML(g.due)}</div>`}</div>`;
        }).join('');
        const rect = cell.getBoundingClientRect();
        detailBox.style.top = `${window.scrollY + rect.top}px`;
        detailBox.style.left = `${window.scrollX + rect.right + 8}px`;
        detailBox.style.display = 'block';
    }
});