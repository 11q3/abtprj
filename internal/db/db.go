package db

import (
	"database/sql"
	"time"
)

func GetDoneTasks(db *sql.DB, start, end time.Time) ([]Task, error) {
	rows, err := db.Query(
		`SELECT name, description, status, done_at
		 FROM tasks
		 WHERE status = 'done' AND done_at >= $1 AND done_at < $2`,
		start, end,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Name, &t.Description, &t.Status, &t.DoneAt); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func GetTodoTasks(db *sql.DB) ([]Task, error) {
	rows, err := db.Query("SELECT name, description, status FROM tasks WHERE status = 'todo'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.Name, &t.Description, &t.Status); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func AddTask(db *sql.DB, name, description string) error {
	_, err := db.Exec(
		"INSERT INTO tasks(name, description, status) VALUES ($1, $2, 'todo')",
		name, description,
	)
	return err
}

func CompleteTask(db *sql.DB, name string) error {
	_, err := db.Exec(
		"UPDATE tasks SET status = 'done', done_at = NOW() WHERE name = $1",
		name,
	)
	return err
}

func GetWorkingStatusForDay(db *sql.DB, date string) ([]WorkSession, error) {
	rows, err := db.Query("SELECT start_time, end_time FROM work_sessions WHERE start_time::date = $1", date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workSessions []WorkSession
	for rows.Next() {
		var ws WorkSession
		if err := rows.Scan(&ws.StartTime, &ws.EndTime); err != nil {
			continue
		}
		workSessions = append(workSessions, ws)
	}
	return workSessions, rows.Err()
}
