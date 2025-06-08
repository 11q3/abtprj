package app

import (
	"abtprj/internal/repository"
	"time"
)

func ConvertRepoTasks(repoTasks []repository.Task) []Task {
	out := make([]Task, len(repoTasks))
	for i, rt := range repoTasks {
		var doneAt *time.Time
		if rt.DoneAt.Valid {
			t := rt.DoneAt.Time
			doneAt = &t
		}
		out[i] = Task{
			Name:        rt.Name,
			Description: rt.Description,
			Status:      rt.Status,
			DoneAt:      doneAt,
		}
	}
	return out
}

func ConvertRepoSessions(repoSessions []repository.WorkSession) []WorkSession {
	out := make([]WorkSession, len(repoSessions))
	for i, rs := range repoSessions {
		var endPtr *time.Time
		if rs.EndTime.Valid {
			t := rs.EndTime.Time
			endPtr = &t
		}
		out[i] = WorkSession{
			ID:        rs.Id,
			StartTime: rs.StartTime,
			EndTime:   endPtr,
		}
	}
	return out
}

func ConvertRepoGoals(repoGoals []repository.Goal) []Goal {
	out := make([]Goal, len(repoGoals))
	for i, goal := range repoGoals {
		out[i] = Goal{
			ID:          goal.Id,
			Name:        goal.Name,
			Description: goal.Description,
			Status:      goal.Status,
			DoneAt:      &goal.DoneAt,
			DueAt:       &goal.DueAt.Time,
		}
	}
	return out
}

/*
type Goal struct {
	Id          int
	Name        string
	Description string
	Status      string
	DoneAt      sql.NullTime
	CreatedAt   time.Time
}
*/
