package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html"
	"log"
	"strings"
	"time"

	"github.com/badoux/checkmail"
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
	ID        uint64    `json:"id" example:"1"`
	Nickname  string    `json:"nickname" example:"Steve"`
	Email     string    `json:"email"  example:"steve@email.com"`
	UserType  string    `json:"user_type"  example:"technician"`
	Password  string    `json:"password,omitempty" example:"password"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-27T20:03:44Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-27T20:03:44Z"`
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

func (u *User) SaveUser(db *sql.DB) (*User, error) {
	err := u.HashPassword()
	if err != nil {
		return &User{}, err
	}

	_, err = db.Exec("INSERT INTO `users` (`nickname`, `email`, `password`) VALUES (?, ?, ?);", &u.Nickname, &u.Email, &u.Password)
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (u *User) FindAllTechnicians(db *sql.DB) (*[]User, error) {
	users := []User{}

	results, err := db.Query("SELECT id, nickname, email FROM users WHERE user_type = ?;", enums.TECHNICIAN)
	if err != nil {
		return &[]User{}, err
	}

	for results.Next() {
		var user User
		err = results.Scan(&user.ID, &user.Nickname, &user.Email)
		if err != nil {
			return &[]User{}, err
		}
		users = append(users, user)
	}

	err = u.RemovePasswords(&users)
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindAllManagers(tx *sql.Tx) (*[]User, error) {
	users := []User{}

	results, err := tx.Query("SELECT id, nickname, email FROM users WHERE user_type = ?;", enums.MANAGER)
	if err != nil {
		return &[]User{}, err
	}

	for results.Next() {
		var user User
		err = results.Scan(&user.ID, &user.Nickname, &user.Email)
		if err != nil {
			return &[]User{}, err
		}
		users = append(users, user)
	}

	err = u.RemovePasswords(&users)
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (u *User) FindUserByID(tx *sql.Tx, uid uint64) (*User, error) {
	err := tx.QueryRow("SELECT id, nickname, email, password, user_type FROM users WHERE id = ?;", uid).Scan(&u.ID, &u.Nickname, &u.Email, &u.Password, &u.UserType)
	switch {
	case err == sql.ErrNoRows:
		return &User{}, errors.New("user not found")
	case err != nil:
		return &User{}, err
	}
	err = u.RemovePassword()
	if err != nil {
		return &User{}, err
	}
	return u, err
}

func (u *User) FindUserByEmail(db *sql.DB, email string) (*User, error) {
	err := db.QueryRow("SELECT id, password FROM users WHERE email = ?;", email).Scan(&u.ID, &u.Password)
	switch {
	case err == sql.ErrNoRows:
		return &User{}, errors.New("user not found")
	case err != nil:
		return &User{}, err
	}
	err = u.RemovePassword()
	if err != nil {
		return &User{}, err
	}
	return u, err
}

func (u *User) UpdateAUser(db *sql.DB, uid uint64) (*User, error) {
	err := u.HashPassword()
	if err != nil {
		log.Fatal(err)
	}

	res, err := db.Exec("UPDATE users SET nickname = ?, email = ?, password = ?, updated_at = ? WHERE id = ?;", &u.Nickname, &u.Email, &u.Password, time.Now(), uid)
	if err != nil {
		return &User{}, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return &User{}, err
	}
	if count > 0 {
		return u, nil
	}
	return &User{}, errors.New("user not found")
}

func (u *User) DeleteAUser(db *sql.DB, uid uint64) (int64, error) {
	res, err := db.Exec("DELETE FROM `users` WHERE id = ?;", uid)
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
