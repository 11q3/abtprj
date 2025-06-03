package handlers

import (
	"abtprj/internal/db"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

const sessionCookieName = "admin_session"

func (h *Handler) requireAdmin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value != "authenticated" {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		handler(w, r)
	}
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if err := h.Templates.ExecuteTemplate(w, "login.html", nil); err != nil {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		}
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusInternalServerError)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	admin, err := db.GetAdminByLogin(h.DB, login)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid login or password", http.StatusBadRequest)
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		log.Printf("Error login: %v", err)
		err := h.Templates.ExecuteTemplate(w, "login.html", map[string]string{
			"Error": "Invalid login or password",
		})
		if err != nil {
			log.Printf("Error executing template: %v", err)
			return
		}
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "authenticated",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}
