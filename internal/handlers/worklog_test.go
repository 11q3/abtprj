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

func TestWorkLogHandler_IsWorkingIndicator(t *testing.T) {
	tmpl := createWorklogTemplate()

	cases := []struct {
		name      string
		isWorking bool
		want      string
	}{
		{"WorkingTrue", true, "IS WORKING:\ntrue"},
		{"WorkingFalse", false, "IS WORKING:\nfalse"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &mockService{
				isWorking:       tc.isWorking,
				tasksForDate:    []app.Task{},
				sessionsForDate: []app.WorkSession{},
			}

			h := &Handler{
				Templates:  tmpl,
				AppService: svc,
			}

			req := httptest.NewRequest(http.MethodGet, "/worklog/", nil)
			rr := httptest.NewRecorder()

			h.WorkLogHandler(rr, req)
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
			}

			body := rr.Body.String()
			if !strings.Contains(body, tc.want) {
				t.Errorf("body = %q; want to contain %q", body, tc.want)
			}
		})
	}
}

func TestWorkLogHandler_OngoingSessionFormatting(t *testing.T) {
	tmpl := createWorklogTemplate()

	startUTC := time.Date(2025, 6, 8, 10, 0, 0, 0, time.UTC)
	sessions := []app.WorkSession{
		{StartTime: startUTC, EndTime: nil},
	}

	svc := &mockService{
		isWorking:       true,
		tasksForDate:    []app.Task{},
		sessionsForDate: sessions,
	}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/worklog/", nil)
	rr := httptest.NewRecorder()

	h.WorkLogHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()
	want := "13:00:00 - (ongoing)"
	if !strings.Contains(body, want) {
		t.Errorf("body = %q; want to contain %q", body, want)
	}
}

func TestWorkLogHandler_DisplayAllSessions(t *testing.T) {
	tmpl := createWorklogTemplate()

	// Two sessions on “today”: one completed, one ongoing.
	// Use UTC times so we can predict Moscow (UTC+3) output.
	start1 := time.Date(2025, time.June, 8, 9, 0, 0, 0, time.UTC)
	end1 := time.Date(2025, time.June, 8, 11, 30, 0, 0, time.UTC)
	start2 := time.Date(2025, time.June, 8, 13, 15, 0, 0, time.UTC)

	sessions := []app.WorkSession{
		{StartTime: start1, EndTime: &end1},
		{StartTime: start2, EndTime: nil},
	}

	svc := &mockService{
		sessionsForDate: sessions,
	}

	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/worklog/", nil)
	rr := httptest.NewRecorder()

	h.WorkLogHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want %d", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	// 09:00 UTC → 12:00 Moscow; 11:30 UTC → 14:30 Moscow
	want1 := "12:00:00 - 14:30:00 (2h30m0s)"
	if !strings.Contains(body, want1) {
		t.Errorf("body missing completed session %q:\n%s", want1, body)
	}

	// 13:15 UTC → 16:15 Moscow, ongoing
	want2 := "16:15:00 - (ongoing)"
	if !strings.Contains(body, want2) {
		t.Errorf("body missing ongoing session %q:\n%s", want2, body)
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
