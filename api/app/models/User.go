package models

import (
	"encoding/json"
	"errors"
	"html"
	"log"
	"strings"
	"time"

	"github.com/badoux/checkmail"
	"github.com/jinzhu/gorm"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"golang.org/x/crypto/bcrypt"
)

type Nickname struct {
	Nickname string `json:"nickname" example:"Steve"`
}

type Email struct {
	Email string `json:"email"  example:"steve@email.com"`
}

type Password struct {
	Password string `json:"password" example:"password"`
}

type User struct {
	ID        uint32    `gorm:"primary_key;auto_increment" json:"id" example:"1"`
	Nickname  string    `gorm:"size:255;not null;unique" json:"nickname" example:"Steve"`
	Email     string    `gorm:"size:100;not null;unique" json:"email"  example:"steve@email.com"`
	UserType  string    `gorm:"type:enum('manager','technician');default:'technician'" json:"user_type"  example:"technician"`
	Password  string    `gorm:"size:100;not null;" json:"password,omitempty" example:"password"`
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at" example:"2023-01-27T20:03:44Z"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at" example:"2023-01-27T20:03:44Z"`
}

func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if u.Nickname == "" {
			return errors.New("required nickname")
		}
		if u.Password == "" {
			return errors.New("required password")
		}
		if u.Email == "" {
			return errors.New("required email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid email")
		}
		if u.UserType != "" {
			return errors.New("invalid argument")
		}

		return nil
	case "login":
		if u.Password == "" {
			return errors.New("required password")
		}
		if u.Email == "" {
			return errors.New("required email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid email")
		}
		return nil

	default:
		if u.Nickname == "" {
			return errors.New("required nickname")
		}
		if u.Password == "" {
			return errors.New("required password")
		}
		if u.Email == "" {
			return errors.New("required email")
		}
		if err := checkmail.ValidateFormat(u.Email); err != nil {
			return errors.New("invalid email")
		}
		if u.UserType != "" {
			return errors.New("invalid argument")
		}
		return nil
	}
}

func (u *User) Prepare() {
	u.ID = 0
	u.Nickname = html.EscapeString(strings.TrimSpace(u.Nickname))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func (u *User) HashPassword() error {
	hashedPassword, err := Hash(u.Password)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (u *User) RemovePassword() error {
	toJson, err := json.Marshal(u)
	if err != nil {
		return err
	}
	toMap := map[string]interface{}{}
	err = json.Unmarshal(toJson, &toMap)
	if err != nil {
		return err
	}
	delete(toMap, "password")
	toJson, err = json.Marshal(toMap)
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(toJson), u)
	if err != nil {
		return err
	}
	return nil
}

func (u *User) RemovePasswords(users *[]User) error {
	toJson, err := json.Marshal(*users)
	if err != nil {
		return err
	}
	toMap := []map[string]interface{}{}
	err = json.Unmarshal(toJson, &toMap)
	if err != nil {
		return err
	}
	for i := range toMap {
		delete(toMap[i], "password")
	}
	toJson, err = json.Marshal(toMap)
	if err != nil {
		return err
	}
	cleanUsers := []User{}
	err = json.Unmarshal([]byte(toJson), &cleanUsers)
	if err != nil {
		return err
	}
	*users = cleanUsers
	return nil
}

func (u *User) SaveUser(db *gorm.DB) (*User, error) {
	err := u.HashPassword()
	if err != nil {
		return &User{}, err
	}
	err = db.Debug().Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) FindAllTechnicians(db *gorm.DB) (*[]User, error) {
	users := []User{}

	err := db.Debug().Model(&User{}).Where("user_type != ?", enums.MANAGER).Limit(100).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	err = u.RemovePasswords(&users)
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindAllManagers(db *gorm.DB) (*[]User, error) {
	users := []User{}

	err := db.Debug().Model(&User{}).Where("user_type = ?", enums.MANAGER).Limit(100).Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	err = u.RemovePasswords(&users)
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserByID(db *gorm.DB, uid uint32) (*User, error) {
	err := db.Debug().Model(User{}).Where("id = ?", uid).Take(&u).Error
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("user not found")
	}
	if err != nil {
		return &User{}, err
	}
	err = u.RemovePassword()
	if err != nil {
		return &User{}, err
	}
	return u, err
}

func (u *User) UpdateAUser(db *gorm.DB, uid uint32) (*User, error) {
	err := u.HashPassword()
	if err != nil {
		log.Fatal(err)
	}
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":   u.Password,
			"nickname":   u.Nickname,
			"email":      u.Email,
			"updated_at": time.Now(),
		},
	)
	if db.Error != nil {
		return &User{}, db.Error
	}
	err = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	err = u.RemovePassword()
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) DeleteAUser(db *gorm.DB, uid uint32) (int64, error) {
	db = db.Debug().Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
