package handlers

import (
	"abtprj/internal/app"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestAdminHandler_NotFound(t *testing.T) {
	tmpl := createAdminTemplate()
	svc := &mockService{}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req1 := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rr1 := httptest.NewRecorder()

	h.AdminHandler(rr1, req1)
	if rr1.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr1.Code, http.StatusNotFound)
	}
}

func TestAdminHandler_OK(t *testing.T) {
	tmpl := createAdminTemplate()

	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mockTodos := []app.Task{
		{
			"testName1",
			"testDescription1",
			"TODO",
			nil,
		},
		{
			"testName2",
			"testDescription2",
			"DONE",
			&t1,
		},
	}

	svc := &mockService{
		todos: mockTodos,
	}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req1 := httptest.NewRequest(http.MethodGet, "/admin/", nil)
	rr1 := httptest.NewRecorder()

	h.AdminHandler(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr1.Code, http.StatusOK)
	}

	body := rr1.Body.String()

	want := "testName1"
	if !strings.Contains(body, want) {
		t.Errorf("handler returned wrong body: got %v want %v", body, want)
	}

	want = "testDescription1"
	if !strings.Contains(body, want) {
		t.Errorf("handler returned wrong body: got %v want %v", body, want)
	}
}

func createAdminTemplate() *template.Template {
	tmpl := template.Must(template.New("admin.html").Parse(`
{{define "admin.html"}}
{{range .Todos}}
NAME:
{{.Name}}
DESCRIPTION:
{{.Description}}
{{end}}
{{end}}
`))

	return tmpl
}
