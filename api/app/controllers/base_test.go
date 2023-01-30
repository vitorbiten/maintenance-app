package controllers

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/joho/godotenv"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/utils"
)

var server = Server{}

func TestMain(m *testing.M) {
	err := godotenv.Load(os.ExpandEnv("../../.env"))
	if err != nil {
		log.Printf("Error getting env %v\n", err)
	}

	Database()
	os.Exit(m.Run())
}

func OnError(err error, text string) {
	if err != nil {
		log.Fatalf(text, err)
	}
}

func Database() {
	DBURL := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"),
	)

	for {
		var err error
		server.DB, err = gorm.Open("mysql", DBURL)
		if err == nil {
			break
		}
		log.Printf("Error connecting to the mysql database: %s", err)
		log.Println("Retrying in 5 seconds...")
		time.Sleep(5 * time.Second)
	}
	log.Println("We are connected to the test mysql daptabase")

	server.DB.Debug().AutoMigrate(&models.User{}, &models.Task{})
}

func RefreshUserTable() error {
	err := server.DB.DropTableIfExists(&models.User{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.User{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed table")
	return nil
}

func SeedOneUser() (models.User, error) {
	err := RefreshUserTable()
	if err != nil {
		log.Fatal(err)
	}
	user := models.User{
		Nickname: "Pet",
		Email:    "pet@gmail.com",
		Password: "password",
		UserType: enums.TECHNICIAN,
	}
	err = user.HashPassword()
	if err != nil {
		return models.User{}, err
	}
	err = server.DB.Model(&models.User{}).Create(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func SeedUsers() ([]models.User, error) {
	var err error
	if err != nil {
		return nil, err
	}

	users := []models.User{
		{
			Nickname: "Victor Reed",
			Email:    "vitor@gmail.com",
			Password: "password",
			UserType: enums.MANAGER,
		},
		{
			Nickname: "Martin Bob",
			Email:    "martin@gmail.com",
			Password: "password",
			UserType: enums.MANAGER,
		},
		{
			Nickname: "Kenny Morris",
			Email:    "kenny@gmail.com",
			Password: "password",
			UserType: enums.TECHNICIAN,
		},
		{
			Nickname: "Denny Morris",
			Email:    "denny@gmail.com",
			Password: "password",
			UserType: enums.TECHNICIAN,
		},
		{
			Nickname: "Billy Joe",
			Email:    "billy@gmail.com",
			Password: "password",
			UserType: enums.TECHNICIAN,
		},
	}

	for i, user := range users {
		err = user.HashPassword()
		if err != nil {
			return []models.User{}, err
		}
		err := server.DB.Model(&models.User{}).Create(&users[i]).Error
		if err != nil {
			return []models.User{}, err
		}
	}
	return users, nil
}

func RefreshUserAndTaskTable() error {
	err := server.DB.DropTableIfExists(&models.User{}, &models.Task{}).Error
	if err != nil {
		return err
	}
	err = server.DB.AutoMigrate(&models.User{}, &models.Task{}).Error
	if err != nil {
		return err
	}
	log.Printf("Successfully refreshed tables")
	return nil
}

func SeedOneUserAndOneTask() (models.User, models.Task, error) {
	err := RefreshUserAndTaskTable()
	if err != nil {
		return models.User{}, models.Task{}, err
	}
	user := models.User{
		Nickname: "Sam Phil",
		Email:    "sam@gmail.com",
		Password: "password",
		UserType: enums.TECHNICIAN,
	}
	err = user.HashPassword()
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
	}
	err = server.DB.Model(&models.User{}).Create(&user).Error
	if err != nil {
		return models.User{}, models.Task{}, err
	}
	task := models.Task{
		ID:       1,
		Summary:  "This is the summary sam",
		AuthorID: 1,
	}
	encryptedTask := task
	err = utils.Encrypt(&encryptedTask.Summary)
	if err != nil {
		log.Fatalf("cannot seed tasks table: %v", err)
	}
	err = server.DB.Model(&models.Task{}).Create(&encryptedTask).Error
	if err != nil {
		return models.User{}, models.Task{}, err
	}
	return user, task, nil
}

func SeedUsersAndTasks() ([]models.User, []models.Task, error) {
	users, err := SeedUsers()
	if err != nil {
		return []models.User{}, []models.Task{}, err
	}

	var tasks = []models.Task{
		{
			Summary:  "Hello world 1",
			AuthorID: 3,
		},
		{
			Summary:  "Hello world 2",
			AuthorID: 4,
		},
	}

	for i := range tasks {
		err = utils.Encrypt(&tasks[i].Summary)
		if err != nil {
			log.Fatalf("cannot seed tasks table: %v", err)
		}
		err = server.DB.Model(&models.Task{}).Create(&tasks[i]).Error
		if err != nil {
			log.Fatalf("cannot seed tasks table: %v", err)
		}
	}
	return users, tasks, nil
}
