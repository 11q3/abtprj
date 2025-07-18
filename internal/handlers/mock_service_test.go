package handlers

import (
	"abtprj/internal/app"
)

type mockService struct {
	taskStats    []app.DayTasksStat
	sessionStats []app.DaySessionsStat

	isWorking       bool
	sessionsForDate []app.WorkSession
	tasksForDate    []app.Task
	todoTasks       []app.Task
	goals           []app.Goal
	todoGoals       []app.Goal
}

func (m *mockService) LoginAdmin(login, password string) error { return nil }
func (m *mockService) AddTask(name, description string) error  { return nil }
func (m *mockService) CompleteTask(name string) error          { return nil }
func (m *mockService) CreateGoal(goal app.Goal) error          { return nil }
func (m *mockService) CompleteGoal(id int) error               { return nil }
func (m *mockService) StartWorkSession() error                 { return nil }
func (m *mockService) EndWorkSession() error                   { return nil }
func (m *mockService) CheckIfAdminExists() (bool, error)       { return false, nil }

func (m *mockService) GetGoals() ([]app.Goal, error) {
	return m.goals, nil
}

func (m *mockService) GetTodoGoals() ([]app.Goal, error) {
	return m.todoGoals, nil
}

func (m *mockService) IsWorking() (bool, error) {
	return m.isWorking, nil
}

func (m *mockService) GetTasksForDate(date string) ([]app.Task, error) {
	return m.tasksForDate, nil
}

func (m *mockService) GetWorkSessionsForDate(date string) ([]app.WorkSession, error) {
	return m.sessionsForDate, nil
}

func (m *mockService) GetTodoTasks() ([]app.Task, error) {
	return m.todoTasks, nil
}

func (m *mockService) GetDayTaskStats(year int) ([]app.DayTasksStat, error) {
	return m.taskStats, nil
}

func (m *mockService) GetDaySessionStats(year int) ([]app.DaySessionsStat, error) {
	return m.sessionStats, nil
}
