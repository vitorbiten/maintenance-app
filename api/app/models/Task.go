package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/vitorbiten/maintenance/api/app/utils"
)

type Summary struct {
	Summary time.Time `json:"summary" example:"Task summary"`
}

type Date struct {
	Date time.Time `json:"date" example:"2023-01-27T20:03:44Z"`
}

type Task struct {
	ID        uint64    `json:"id" example:"1"`
	Summary   string    `json:"summary" example:"Task summary"`
	AuthorID  uint64    `json:"author_id" example:"3"`
	Date      time.Time `json:"date" example:"2023-01-27T20:03:44Z"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-27T20:03:44Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-27T20:03:44Z"`
}

func (t *Task) Validate() error {
	if t.Summary == "" {
		return errors.New("required summary")
	}
	if len(t.Summary) <= 5 {
		return errors.New("summary min length is 5 characters")
	}
	if len(t.Summary) > 2500 {
		return errors.New("summary max length is 2500 characters")
	}
	return nil
}

func (t *Task) Prepare() error {
	err := utils.Encrypt(&t.Summary)
	if err != nil {
		return err
	}
	now := time.Now()
	if t.Date.IsZero() {
		t.Date = now
	}
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

func (t *Task) DecryptSummary() error {
	err := utils.Decrypt(&t.Summary)
	if err != nil {
		return err
	}
	return nil
}

func (t *Task) DecryptSummaries(tasks *[]Task) error {
	for _, task := range *tasks {
		err := utils.Decrypt(&task.Summary)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Task) SaveTask(tx *sql.Tx) (int64, error) {
	res, err := tx.Exec("INSERT INTO `tasks` (`summary`, `date`, `author_id`) VALUES (?, ?, ?);", &t.Summary, &t.Date, &t.AuthorID)
	if err != nil {
		return 0, err
	}
	err = t.DecryptSummary()
	if err != nil {
		return 0, err
	}
	lastInsertedId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertedId, nil
}

func (t *Task) FindAllTasks(db *sql.DB) (*[]Task, error) {
	tasks := []Task{}

	results, err := db.Query("SELECT id, summary, date, author_id, created_at, updated_at FROM tasks;")
	if err != nil {
		return &[]Task{}, err
	}

	for results.Next() {
		var task Task
		err = results.Scan(&task.ID, &task.Summary, &task.Date, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return &[]Task{}, err
		}
		tasks = append(tasks, task)
	}

	err = t.DecryptSummaries(&tasks)
	if err != nil {
		return &[]Task{}, err
	}
	return &tasks, nil
}

func (t *Task) FindTasksByAuthorID(db *sql.DB, tid uint64) (*[]Task, error) {
	tasks := []Task{}

	results, err := db.Query("SELECT id, summary, date, author_id, created_at, updated_at FROM tasks WHERE author_id = ?;", tid)
	if err != nil {
		return &[]Task{}, err
	}

	for results.Next() {
		var task Task
		err = results.Scan(&task.ID, &task.Summary, &task.Date, &task.AuthorID, &task.CreatedAt, &task.UpdatedAt)
		if err != nil {
			return &[]Task{}, err
		}
		tasks = append(tasks, task)
	}

	err = t.DecryptSummaries(&tasks)
	if err != nil {
		return &[]Task{}, err
	}
	return &tasks, nil
}

func (t *Task) FindTaskByID(db *sql.DB, tid uint64) (*Task, error) {
	err := db.QueryRow("SELECT id, summary, date, author_id, created_at, updated_at FROM tasks WHERE id = ?;", tid).Scan(&t.ID, &t.Summary, &t.Date, &t.AuthorID, &t.CreatedAt, &t.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		return &Task{}, errors.New("task not found")
	case err != nil:
		return &Task{}, err
	}
	err = t.DecryptSummary()
	if err != nil {
		return &Task{}, err
	}
	return t, err
}

func (t *Task) UpdateATask(db *sql.DB, tid uint64) (*Task, error) {
	res, err := db.Exec("UPDATE tasks SET summary = ?, date = ?, updated_at = ? WHERE id = ?;", &t.Summary, &t.Date, time.Now(), tid)
	if err != nil {
		return &Task{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return &Task{}, err
	}
	if count > 0 {
		err = t.DecryptSummary()
		if err != nil {
			return &Task{}, err
		}
		return t, nil
	}
	return &Task{}, nil
}

func (t *Task) DeleteATask(db *sql.DB, tid uint64) (int64, error) {
	res, err := db.Exec("DELETE FROM `tasks` WHERE id = ?;", tid)
	if err != nil {
		return 0, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 1, nil
	}
	return 0, nil
}
