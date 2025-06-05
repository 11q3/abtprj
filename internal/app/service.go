package app

import (
	"abtprj/internal/repository"
	"abtprj/internal/utils"
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

	CheckIfAdminExists() (bool, error)
	GetTodoTasks() ([]Task, error)

	GetDayTaskStats(year int) ([]DayTasksStat, error)
	GetDaySessionStats(year int) ([]DaySessionsStat, error)
}

type DefaultAppService struct {
	DB *sql.DB
}

func NewDefaultAppService(db *sql.DB) *DefaultAppService {
	return &DefaultAppService{DB: db}
}

type DayTasksStat struct {
	Date  string // "2023-06-01"
	Count int    // how many tasks completed that day
	Level int    // shade level 0–4
	Row   int    // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col   int    // grid‐column (2–54): week index + 2
}

type DaySessionsStat struct {
	Date       string // "2023-06-01"
	SessionDur time.Duration
	Level      int // 0–4 shade
	Row        int // grid‐row (2–8): 2=Sunday, 3=Monday … 8=Saturday
	Col        int // grid‐column (2–54): week index + 2
}

func (s *DefaultAppService) AddTask(name string, description string) error {
	err := repository.AddTask(s.DB, name, description)
	if err != nil {
		return err
	}
	return nil
}

func (s *DefaultAppService) CompleteTask(name string) error {
	err := repository.CompleteTask(s.DB, name)
	if err != nil {
		return err
	}
	return nil
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
	start, end, err := utils.ParseDateRange(date)
	if err != nil {
		return nil, err
	}
	repoTasks, err := repository.GetDoneTasks(s.DB, start, end)
	if err != nil {
		return nil, err
	}
	return ConvertRepoTasks(repoTasks), nil
}

func (s *DefaultAppService) GetWorkSessionsForDate(date string) ([]WorkSession, error) {
	start, end, err := utils.ParseDateRange(date)
	if err != nil {
		return nil, err
	}
	repoSessions, err := repository.GetWorkingSessionsForDay(s.DB, start, end)
	if err != nil {
		return nil, err
	}
	return ConvertRepoSessions(repoSessions), nil
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

		case stats[idx].SessionDur > 1*time.Second:
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
	err := repository.StartWorkSession(s.DB)
	if err != nil {
		log.Printf("startWorkSession exec error: %v", err)
		return err
	}
	return nil
}

func (s *DefaultAppService) EndWorkSession() error {
	err := repository.EndWorkSession(s.DB)
	if err != nil {
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

func generateEmptyDayStats() []DayTasksStat {
	days := make([]DayTasksStat, 0, 52*7)
	for week := 1; week <= 52; week++ {
		for dow := 0; dow < 7; dow++ {
			days = append(days, DayTasksStat{"", 0, 0, dow + 2, week + 1})
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
