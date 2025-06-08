package app

import (
	"abtprj/internal/repository"
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type AppService interface {
	LoginAdmin(login, password string) error
	AddTask(name, description string) error
	CompleteTask(name string) error
	GetTasksForDate(date string) ([]Task, error)
	GetWorkSessionsForDate(date string) ([]WorkSession, error)
	StartWorkSession() error
	EndWorkSession() error
	GetGoals() ([]Goal, error)
	GetTodoGoals() ([]Goal, error)
	CompleteGoal(id int) error
	CreateGoal(goal Goal) error

	IsWorking() (bool, error)

	CheckIfAdminExists() (bool, error)
	GetTodoTasks() ([]Task, error)

	GetDayTaskStats(year int) ([]DayTasksStat, error)
	GetDaySessionStats(year int) ([]DaySessionsStat, error)
}

type DefaultAppService struct {
	DB  *sql.DB
	loc *time.Location
}

func NewDefaultAppService(db *sql.DB) *DefaultAppService {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		loc = time.UTC
	}
	return &DefaultAppService{DB: db, loc: loc}
}

type DayTasksStat struct {
	Date  string // "2023-06-01"
	Count int    // how many tasks completed that day
	Level int    // shade level 0–4
	Row   int    // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col   int    // grid‐column (2–54): week index + 2
	Goals []Goal
}

type DaySessionsStat struct {
	Date       string // "2023-06-01"
	SessionDur time.Duration
	Level      int // 0–4 shade
	Row        int // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col        int // grid‐column (2–54): week index + 2
}

func (s *DefaultAppService) AddTask(name string, description string) error {
	return repository.AddTask(s.DB, name, description)
}

func (s *DefaultAppService) CompleteTask(name string) error {
	return repository.CompleteTask(s.DB, name)
}

func (s *DefaultAppService) LoginAdmin(login, password string) error {
	admin, err := repository.GetAdminByLogin(s.DB, login)
	if err != nil {
		return err
	}
	if bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)) != nil {
		return errors.New("invalid password")
	}
	return nil
}

func (s *DefaultAppService) GetTasksForDate(date string) ([]Task, error) {
	day, err := time.ParseInLocation("2006-01-02", date, s.loc)
	if err != nil {
		return nil, err
	}
	startUTC := day.UTC()
	endUTC := day.Add(24 * time.Hour).UTC()

	repoTasks, err := repository.GetDoneTasks(s.DB, startUTC, endUTC)
	if err != nil {
		return nil, err
	}
	tasks := ConvertRepoTasks(repoTasks)
	for i := range tasks {
		if tasks[i].DoneAt != nil {
			t := tasks[i].DoneAt.In(s.loc)
			tasks[i].DoneAt = &t
		}
	}
	return tasks, nil
}

func (s *DefaultAppService) GetWorkSessionsForDate(date string) ([]WorkSession, error) {
	day, err := time.ParseInLocation("2006-01-02", date, s.loc)
	if err != nil {
		return nil, err
	}
	startUTC := day.UTC()
	endUTC := day.Add(24 * time.Hour).UTC()

	repoSessions, err := repository.GetWorkingSessionsForDay(s.DB, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	sessions := ConvertRepoSessions(repoSessions)
	for i := range sessions {
		sessions[i].StartTime = sessions[i].StartTime.In(s.loc)
		if sessions[i].EndTime != nil {
			t2 := sessions[i].EndTime.In(s.loc)
			sessions[i].EndTime = &t2
		}
	}
	return sessions, nil
}

func (s *DefaultAppService) GetDayTaskStats(year int) ([]DayTasksStat, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)
	tasks, err := repository.GetDoneTasks(s.DB, start, end)
	if err != nil {
		return nil, err
	}

	empty := generateEmptyDayStats()
	stats := make([]DayTasksStat, len(empty))
	copy(stats, empty)
	for _, task := range tasks {
		if !task.DoneAt.Valid {
			continue
		}
		d := task.DoneAt.Time
		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1
		dowIdx := (int(d.Weekday()) + 6) % 7
		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			continue
		}
		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].Count++
		lvl := stats[idx].Count
		if lvl > 4 {
			lvl = 4
		}
		stats[idx].Level = lvl
	}

	empty = generateEmptyDayStats()
	copy(stats, stats)
	repoGoals, err := repository.GetGoals(s.DB)
	goals := ConvertRepoGoals(repoGoals)
	if err != nil {
		return nil, err
	}
	for _, goal := range goals {
		if goal.DueAt == nil {
			continue
		}
		d := goal.DueAt
		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1
		dowIdx := (int(d.Weekday()) + 6) % 7
		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			continue
		}
		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].Goals = append(stats[idx].Goals, goal)
	}
	return stats, nil
}

func (s *DefaultAppService) GetDaySessionStats(year int) ([]DaySessionsStat, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, time.December, 31, 23, 59, 59, 0, time.UTC)

	sessions, err := repository.GetWorkingSessions(s.DB, start, end)
	if err != nil {
		return nil, err
	}

	empty := generateEmptySessionStats()
	stats := make([]DaySessionsStat, len(empty))
	copy(stats, empty)

	for _, sess := range sessions {
		if !sess.EndTime.Valid {
			continue
		}
		d := sess.EndTime.Time

		_, isoWeek := d.ISOWeek()
		weekIdx := isoWeek - 1

		// convert Go’s Sunday=0…Saturday=6 → Monday=0…Sunday=6
		dowIdx := (int(d.Weekday()) + 6) % 7

		idx := weekIdx*7 + dowIdx
		if idx < 0 || idx >= len(stats) {
			continue
		}

		stats[idx].Date = d.Format("2006-01-02")
		stats[idx].SessionDur += sess.EndTime.Time.Sub(sess.StartTime)
		switch {
		case stats[idx].SessionDur > 8*time.Hour:
			stats[idx].Level = 4
		case stats[idx].SessionDur > 4*time.Hour:
			stats[idx].Level = 3
		case stats[idx].SessionDur > 2*time.Hour:
			stats[idx].Level = 2
		case stats[idx].SessionDur > 0:
			stats[idx].Level = 1
		default:
			stats[idx].Level = 0
		}
		if stats[idx].Level > 4 {
			stats[idx].Level = 4
		}
	}

	return stats, nil
}

func (s *DefaultAppService) StartWorkSession() error {
	if err := repository.StartWorkSession(s.DB); err != nil {
		log.Printf("startWorkSession exec error: %v", err)
		return err
	}
	return nil
}

func (s *DefaultAppService) IsWorking() (bool, error) {
	isWorking, _, err := repository.CheckIfActiveSessions(s.DB)
	if err != nil {
		return false, err
	}
	return isWorking, nil
}

func (s *DefaultAppService) EndWorkSession() error {
	if err := repository.EndWorkSession(s.DB); err != nil {
		log.Printf("endWorkSession exec error: %v", err)
		return err
	}
	return nil
}

func (s *DefaultAppService) GetTodoTasks() ([]Task, error) {
	repoTasks, err := repository.GetTodoTasks(s.DB)
	if err != nil {
		log.Printf("GetTodoTasks exec error: %v", err)
		return nil, err
	}
	return ConvertRepoTasks(repoTasks), nil
}

func (s *DefaultAppService) CheckIfAdminExists() (bool, error) {
	exists, err := repository.CheckIfAdminExists(s.DB)
	if err != nil {
		log.Printf("CheckIfAdminExists exec error: %v", err)
		return false, err
	}
	return exists, nil
}

func (s *DefaultAppService) GetGoals() ([]Goal, error) {
	goals, err := repository.GetGoals(s.DB)
	if err != nil {
		log.Printf("GetGoal exec error: %v", err)
		return nil, err
	}
	return ConvertRepoGoals(goals), err
}

func (s *DefaultAppService) GetTodoGoals() ([]Goal, error) {
	todoGoals, err := repository.GetTodoGoals(s.DB)
	if err != nil {
		log.Printf("GetTodoGoal exec error: %v", err)
		return nil, err
	}
	return ConvertRepoGoals(todoGoals), err
}

func (s *DefaultAppService) CompleteGoal(id int) error {
	err := repository.CompleteGoal(s.DB, id)
	if err != nil {
		log.Printf("CompleteGoal exec error: %v", err)
		return err
	}
	return nil
}

func (s *DefaultAppService) CreateGoal(goal Goal) error {
	err := repository.CreateGoal(s.DB, goal.Name, goal.Description, *goal.DueAt)
	if err != nil {
		log.Printf("CreateGoal exec error: %v", err)
	}
	return nil
}

func generateEmptyDayStats() []DayTasksStat {
	days := make([]DayTasksStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DayTasksStat{"", 0, 0, dow + 2, week + 1, nil})
		}
	}
	return days
}

func generateEmptySessionStats() []DaySessionsStat {
	days := make([]DaySessionsStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DaySessionsStat{"", time.Second * 0, 0, dow + 2, week + 1})
		}
	}
	return days
}
