package db

type Task struct {
	Name        string
	Description string
	Status      string
}

type WorkSession struct {
	Date      string
	StartTime string
	EndTime   string
	Duration  string
	Name      string
	Status    string
	Tasks     []Task
}
