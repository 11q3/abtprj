package app

import "time"

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
