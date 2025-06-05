package handlers

import (
	"abtprj/internal/app"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"text/template"
	"time"
)

type mockService struct {
	taskStats     []app.DayTasksStat
	sessionStats  []app.DaySessionsStat
	errOnTasks    bool
	errOnSessions bool
}

func (m *mockService) LoginAdmin(login, password string) error                       { return nil }
func (m *mockService) AddTask(name, description string) error                        { return nil }
func (m *mockService) CompleteTask(name string) error                                { return nil }
func (m *mockService) GetTasksForDate(date string) ([]app.Task, error)               { return nil, nil }
func (m *mockService) GetWorkSessionsForDate(date string) ([]app.WorkSession, error) { return nil, nil }
func (m *mockService) StartWorkSession() error                                       { return nil }
func (m *mockService) EndWorkSession() error                                         { return nil }
func (m *mockService) CheckIfAdminExists() (bool, error)                             { return false, nil }
func (m *mockService) GetTodoTasks() ([]app.Task, error)                             { return nil, nil }

func (m *mockService) GetDayTaskStats(year int) ([]app.DayTasksStat, error) {
	if m.errOnTasks {
		return nil, errors.New("simulated task‐stats failure")
	}
	return m.taskStats, nil
}

func (m *mockService) GetDaySessionStats(year int) ([]app.DaySessionsStat, error) {
	if m.errOnSessions {
		return nil, errors.New("simulated session‐stats failure")
	}
	return m.sessionStats, nil
}

func TestStatsHandler_NotFound(t *testing.T) {
	tmpl := template.Must(template.New("stats.html").Parse(`{{define "stats.html"}}OK{{end}}`))
	svc := &mockService{}
	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req1 := httptest.NewRequest(http.MethodGet, "/stats", nil) //creates a fake http request object
	rr1 := httptest.NewRecorder()                              //creates implementation that records what handler writes into it
	h.StatsHandler(rr1, req1)
	if rr1.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr1.Code, http.StatusNotFound)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/stats", nil)
	rr2 := httptest.NewRecorder()
	h.StatsHandler(rr2, req2)
	if rr2.Code != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", rr2.Code, http.StatusNotFound)
	}
}

func TestStatsHandler_OK(t *testing.T) {
	tmpl := template.Must(template.New("stats.html").Parse(`
{{define "stats.html"}}
TASKS:
{{- range .TaskContributions}}
{{.Date}}|{{.Count}}; 
{{- end}}
SESSIONS:
{{- range .SessionContributions}}
{{.Date}}|{{.SessionDur}}; 
{{- end}}
{{end}}
`))

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

	svc := &mockService{
		taskStats:    mockTasks,
		sessionStats: mockSessions,
	}
	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/stats/", nil)
	rr := httptest.NewRecorder()
	h.StatsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}

	body := rr.Body.String()

	for _, ts := range mockTasks {
		want := ts.Date + "|" + strings.TrimSpace(strings.TrimSpace(strings.TrimPrefix(strings.TrimSuffix(time.Duration(ts.Count).String(), "s"), "")))
		want = ts.Date + "|" + string(rune('0'+ts.Count))
		if !strings.Contains(body, want) {
			t.Errorf("handler returned wrong response body: got %v want %v", body, want)
		}
	}

	for _, ss := range mockSessions {
		durStr := ss.SessionDur.String()
		want := ss.Date + "|" + durStr
		if !strings.Contains(body, want) {
			t.Errorf("expected body to contain sessions entry %q, but it did not", want)
		}
	}
}

func TestStatsHandler_SessionStatsError(t *testing.T) {
	tmpl := template.Must(template.New("stats.html").Parse(`{{define "stats.html"}}OK{{end}}`))
	svc := &mockService{errOnSessions: true}
	h := &Handler{
		Templates:  tmpl,
		AppService: svc,
	}

	req := httptest.NewRequest(http.MethodGet, "/stats/", nil)
	rr := httptest.NewRecorder()
	h.StatsHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusInternalServerError)
	}
	if !strings.Contains(rr.Body.String(), "failed to load stats") {
		t.Errorf("handler returned wrong response body: got %v want %v", rr.Body.String(), "failed to load stats")
	}
}
