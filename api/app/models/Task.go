package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/vitorbiten/maintenance/api/app/utils"
)

type Summary struct {
	Summary time.Time `json:"summary" example:"Task summary"`
}

type Date struct {
	Date time.Time `json:"date" example:"2023-01-27T20:03:44Z"`
}

type Task struct {
	ID        uint64    `gorm:"primary_key;auto_increment" json:"id" example:"1"`
	Summary   string    `gorm:"type:text;not null;" json:"summary" example:"Task summary"`
	AuthorID  uint32    `gorm:"not null" json:"author_id" example:"3"`
	Date      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"date" example:"2023-01-27T20:03:44Z"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at" example:"2023-01-27T20:03:44Z"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at" example:"2023-01-27T20:03:44Z"`
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
	t.ID = 0
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

func (t *Task) SaveTask(db *gorm.DB) (*Task, error) {
	err := db.Debug().Model(&Task{}).Create(&t).Error
	if err != nil {
		return &Task{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Error
		if err != nil {
			return &Task{}, err
		}
		err = t.DecryptSummary()
		if err != nil {
			return &Task{}, err
		}
	}
	return t, nil
}

func (t *Task) FindAllTasks(db *gorm.DB) (*[]Task, error) {
	tasks := []Task{}

	err := db.Debug().Model(&Task{}).Limit(100).Find(&tasks).Error
	if err != nil {
		return &[]Task{}, err
	}
	if len(tasks) > 0 {
		for i := range tasks {
			err := db.Debug().Model(&User{}).Where("id = ?", tasks[i].AuthorID).Error
			if err != nil {
				return &[]Task{}, err
			}
		}
	}
	err = t.DecryptSummaries(&tasks)
	if err != nil {
		return &[]Task{}, err
	}
	return &tasks, nil
}

func (t *Task) FindTasksByAuthorID(db *gorm.DB, pid uint64) (*[]Task, error) {
	tasks := []Task{}

	err := db.Debug().Model(&Task{}).Limit(100).Where("author_id = ?", pid).Find(&tasks).Error
	if err != nil {
		return &[]Task{}, err
	}
	err = t.DecryptSummaries(&tasks)
	if err != nil {
		return &[]Task{}, err
	}
	return &tasks, nil
}

func (t *Task) FindTaskByID(db *gorm.DB, pid uint64) (*Task, error) {
	err := db.Debug().Model(&Task{}).Where("id = ?", pid).Take(&t).Error
	if gorm.IsRecordNotFoundError(err) {
		return &Task{}, errors.New("task not found")
	}
	if err != nil {
		return &Task{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Error
		if err != nil {
			return &Task{}, err
		}
	}
	err = t.DecryptSummary()
	if err != nil {
		return &Task{}, err
	}
	return t, nil
}

func (t *Task) UpdateATask(db *gorm.DB, pid uint64) (*Task, error) {
	db = db.Debug().Model(&Task{}).Where("id = ?", pid).Take(&Task{}).UpdateColumns(
		map[string]interface{}{
			"date":       t.Date,
			"summary":    t.Summary,
			"updated_at": time.Now(),
		},
	)
	err := db.Debug().Model(&Task{}).Where("id = ?", pid).Take(&t).Error
	if err != nil {
		return &Task{}, err
	}
	if t.ID != 0 {
		err = db.Debug().Model(&User{}).Where("id = ?", t.AuthorID).Error
		if err != nil {
			return &Task{}, err
		}
		err := t.DecryptSummary()
		if err != nil {
			return &Task{}, err
		}
	}
	return t, nil
}

func (t *Task) DeleteATask(db *gorm.DB, pid uint64) (int64, error) {
	db = db.Debug().Model(&Task{}).Where("id = ?", pid).Take(&Task{}).Delete(&Task{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Task not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
