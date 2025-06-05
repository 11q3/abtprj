package handlers

import (
	"net/http"
)

func (h *Handler) MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if err := h.Templates.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, "failed to render index.html", http.StatusInternalServerError)
	}
}
