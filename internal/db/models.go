package db

import (
	"database/sql"
	"time"
)

type Task struct {
	Name        string
	Description string
	Status      string
	CreatedAt   time.Time
	DoneAt      sql.NullTime
}

type WorkSession struct {
	Date      string // YYYY-MM-D
	StartTime string
	EndTime   string
	Duration  string
	Name      string
	Status    string
	Tasks     []Task
}
