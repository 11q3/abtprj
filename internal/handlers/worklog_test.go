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

func TestWorkLogHandler_NotFound(t *testing.T) {
	tmpl := createWorklogTemplate()
	svc := &mockService{}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req1 := httptest.NewRequest(http.MethodGet, "/worklog", nil)
	rr1 := httptest.NewRecorder()

	h.WorkLogHandler(rr1, req1)
	if rr1.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr1.Code, http.StatusNotFound)
	}
}

func TestWorkLogHandler_OK(t *testing.T) {
	tmpl := createWorklogTemplate()

	mockTasks := []app.DayTasksStat{
		{
			Date:  "2025-01-01",
			Count: 3,
			Level: 2,
			Row:   4,
			Col:   3,
		},
		{
			Date:  "2025-01-02",
			Count: 0,
			Level: 0,
			Row:   5,
			Col:   3,
		},
	}

	mockSessions := []app.DaySessionsStat{
		{
			Date:       "2025-01-01",
			SessionDur: 2 * time.Hour,
			Level:      1,
			Row:        4,
			Col:        3,
		},
		{
			Date:       "2025-01-02",
			SessionDur: 5 * time.Hour,
			Level:      3,
			Row:        5,
			Col:        3,
		},
	}

	t1 := time.Date(2025, time.January, 23, 0, 5, 0, 0, time.UTC)
	tasksForDate := []app.Task{
		{"testTask1",
			"testDescription1",
			"DONE",
			&t1,
		},
		{"testTask2",
			"testDescription2",
			"TODO",
			&t1,
		},
	}

	t1Start := time.Date(2025, time.January, 23, 9, 0, 0, 0, time.UTC)
	t1End := time.Date(2025, time.January, 23, 11, 0, 0, 0, time.UTC)

	t2Start := time.Date(2025, time.January, 23, 13, 0, 0, 0, time.UTC)
	t2End := time.Date(2025, time.January, 23, 14, 30, 0, 0, time.UTC)

	sessionsForDate := []app.WorkSession{
		{
			ID:        1001,
			StartTime: t1Start,
			EndTime:   &t1End,
		},
		{
			ID:        1002,
			StartTime: t2Start,
			EndTime:   &t2End,
		},
	}

	svc := &mockService{
		taskStats:       mockTasks,
		sessionStats:    mockSessions,
		tasksForDate:    tasksForDate,
		sessionsForDate: sessionsForDate,
	}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req1 := httptest.NewRequest(http.MethodGet, "/worklog/", nil)
	rr1 := httptest.NewRecorder()

	h.WorkLogHandler(rr1, req1)

	if rr1.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr1.Code, http.StatusOK)
	}

	body := rr1.Body.String()

	want := "12:00:00 - 14:00:00 (2h0m0s)"
	if !strings.Contains(body, want) {
		t.Errorf("expected session entry %q, got:\n%s", want, body)
	}

	want = "16:00:00 - 17:30:00 (1h30m0s)"
	if !strings.Contains(body, want) {
		t.Errorf("expected session entry %q, got:\n%s", want, body)
	}

	if !strings.Contains(body, "3h30m0s") {
		t.Errorf("expected total session duration “3h30m0s”, got:\n%s", body)
	}
}

func createWorklogTemplate() *template.Template {
	tmpl := template.Must(template.New("worklog.html").Parse(`
{{define "worklog.html"}}
IS WORKING:
{{ .IsWorking }}
ALL SESSIONS:
{{ range .AllSessions }}
{{.}}
{{end}}
TOTAL SESSION DURATION:
{{.TotalSessionDur}}
TASKS:
{{range .Dones}}
NAME: 
{{.Name}}
DESCRIPTION:
{{.Description}}
DONE AT:
{{ if .DoneAt }}
  {{ .DoneAt.Format "2006-01-02 15:04:05" }}
{{ else }}
{{end}}
{{end}}
{{end}}
`))

	return tmpl
}
