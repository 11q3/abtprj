package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
)

func TestRequireAdmin_NoCookie_Redirects(t *testing.T) {
	h := &Handler{}
	innerCalled := false
	inner := func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.Write([]byte("SECRET"))
	}

	wrapped := h.requireAdmin(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/secret", nil)
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if innerCalled {
		t.Fatal("expected handler to not call inner handler when no cookie set")
	}
	if rr.Code != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got: %s", loc)
	}
}

func TestRequireAdmin_WrongCookie_Redirects(t *testing.T) {
	h := &Handler{}
	innerCalled := false
	inner := func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
	}

	wrapped := h.requireAdmin(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/secret", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "wrong"})

	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if innerCalled {
		t.Fatal("expected handler to not call inner handler when no cookie set")
	}
	if rr.Code != http.StatusFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusFound)
	}
	if loc := rr.Header().Get("Location"); loc != "/login" {
		t.Errorf("expected redirect to /login, got: %s", loc)
	}

}

func TestRequireAdmin_ValidCookie_Allows(t *testing.T) {
	h := &Handler{}
	innerCalled := false
	inner := func(w http.ResponseWriter, r *http.Request) {
		innerCalled = true
		w.Write([]byte("OK"))
	}

	wrapped := h.requireAdmin(inner)

	req := httptest.NewRequest(http.MethodGet, "/admin/secret", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "authenticated"})
	rr := httptest.NewRecorder()

	wrapped(rr, req)

	if !innerCalled {
		t.Fatal("expected handler to call inner handler when valid cookie set")
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status: %v1, got: %v2", http.StatusOK, rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("expected body: %s, got: %s", "OK", rr.Body.String())
	}
}

func TestLoginPage_RendersTemplate(t *testing.T) {
	tmpl := createLoginTemplate()
	h := &Handler{Templates: tmpl}

	req := httptest.NewRequest(http.MethodGet, "/login", nil)
	rec := httptest.NewRecorder()

	h.LoginPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status: %d, got: %d", http.StatusOK, rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "LOGIN PAGE") {
		t.Errorf("expected body: %s, got: %s", "LOGIN PAGE", body)
	}
}

func TestLoginPage_Success_SetsCookieAndRedirects(t *testing.T) {
	tmpl := createLoginTemplate()
	svc := &mockService{}
	h := &Handler{Templates: tmpl, AppService: svc}

	form := strings.NewReader("login=admin&password=secret")
	req := httptest.NewRequest(http.MethodPost, "/login", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	h.HandleLogin(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status: %d, got: %d", http.StatusSeeOther, rr.Code)
	}
	if loc := rr.Header().Get("Location"); loc != "/admin/" {
		t.Errorf("expected redirect to /admin/, got: %s", loc)
	}

	cookies := rr.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == sessionCookieName && cookie.Value == "authenticated" {
			found = true
		}
	}
	if !found {
		t.Error("expected cookie &q=authenticated, but it wasn't set", sessionCookieName)
	}
}

func createLoginTemplate() *template.Template {
	return template.Must(template.New("login.html").Parse(`
{{define "login.html"}}
LOGIN PAGE
{{end}}
`))
}
