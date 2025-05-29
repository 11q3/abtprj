package db

import (
	"database/sql"
	"errors"
	"log"
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
	isActive, session, err := checkIfActiveSessions(db)
	if !isActive || session == nil {
		log.Printf("attempting to end a task without active session: %v", err)
		return err
	}
	result, err := db.Exec(
		"UPDATE tasks SET status = 'done', done_at = NOW(), session_id = $1 WHERE name = $2",
		session.Id,
		name,
	)

	log.Printf("sql.Result: %#v", result)
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

func StartWorkSession(db *sql.DB) error {
	isActive, _, err := checkIfActiveSessions(db)
	if isActive {
		log.Printf("Attepmpting to create another worksession, while active sessions exist %v", err)
		return err
	}

	_, err = db.Exec("INSERT INTO work_sessions(start_time) VALUES ($1)", time.Now())
	if err != nil {
		log.Printf("Error inserting work session: %v", err)
		return err
	}
	return err
}

func EndWorkSession(db *sql.DB) error {
	isActive, s, err := checkIfActiveSessions(db)
	if !isActive {
		log.Printf("Attepmpting to end worksession, while active sessions does not exist %v", err)
		return err
	}

	_, err = db.Exec("UPDATE work_sessions set end_time = $1 WHERE ID=$2", time.Now(), s.Id)
	if err != nil {
		log.Printf("Error inserting work session: %v", err)
		return err
	}
	return nil
}

func checkIfActiveSessions(db *sql.DB) (bool, *WorkSession, error) {
	row := db.QueryRow("SELECT * FROM work_sessions WHERE end_time IS NULL")

	var ws WorkSession
	err := row.Scan(&ws.Id, &ws.StartTime, &ws.Duration, &ws.EndTime, &ws.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil, nil
		}
		log.Printf("Error scanning work session: %v", err)
		return false, &ws, err
	}

	return true, &ws, nil
}
