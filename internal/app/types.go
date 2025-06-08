package app

import (
	"database/sql"
	"time"
)

type Task struct {
	Name        string
	Description string
	Status      string
	DoneAt      *time.Time
}

type WorkSession struct {
	ID        int
	StartTime time.Time
	EndTime   *time.Time
}

type Goal struct {
	ID          int
	Name        string
	Description string
	Status      string
	DoneAt      *sql.NullTime
	DueAt       *time.Time
}
