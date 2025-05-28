package handlers

import (
	"log"
	"net/http"

	"abtprj/internal/db"
	"abtprj/internal/utils"
)

func (h *Handler) WorkLogHandler(w http.ResponseWriter, r *http.Request) {
	start, end, err := utils.ParseDateRange(r.URL.Query().Get("date"))
	if err != nil {
		http.Error(w, "invalid date format", http.StatusBadRequest)
		return
	}

	tasks, err := db.GetDoneTasks(h.DB, start, end)
	if err != nil {
		log.Printf("worklog query error: %v", err)
		http.Error(w, "db query error", http.StatusInternalServerError)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "worklog.html", struct{ Tasks []db.Task }{tasks}); err != nil {
		log.Printf("template exec error: %v", err)
	}
}
