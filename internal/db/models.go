package db

import (
	"database/sql"
	"time"
)

type Task struct {
	Id          int
	Name        string
	Description string
	Status      string
	DoneAt      sql.NullTime
	CreatedAt   time.Time
}

type WorkSession struct {
	Id        int
	Date      string // YYYY-MM-D
	StartTime time.Time
	EndTime   sql.NullTime
	Duration  sql.NullString
	Name      string
	Status    string
	Tasks     []Task
	CreatedAt time.Time
}
