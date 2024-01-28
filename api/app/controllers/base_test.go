package controllers

import (
	"log"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/utils"
)

func TestMain(m *testing.M) {
	err := godotenv.Load(os.ExpandEnv("../../.env"))
	if err != nil {
		log.Printf("Error getting env %v\n", err)
	}

	gin.SetMode(gin.TestMode)

	Database()
	os.Exit(m.Run())
}

func OnError(err error, text string) {
	if err != nil {
		log.Fatalf(text, err)
	}
}

func SetupRouter() *gin.Engine {
	router := gin.Default()
	InitializeRoutes(router)
	return router
}

func Database() {
	adapters.LoadTestDatabase()
	err := MigrateDB()
	if err != nil {
		log.Fatalf("cannot migrate db: %s", err)
	}
}

func MigrateDB() error {
	var err error
	_, err = adapters.DB.Exec("DROP DATABASE IF EXISTS `maintenance_api_test`")
	if err != nil {
		log.Fatalf("cannot drop users table: %s", err)
	}
	_, err = adapters.DB.Exec("CREATE DATABASE IF NOT EXISTS `maintenance_api_test`")
	if err != nil {
		log.Fatalf("cannot create database: %s", err)
	}
	_, err = adapters.DB.Exec("USE `maintenance_api_test`")
	if err != nil {
		log.Fatalf("cannot use database: %s", err)
	}
	_, err = adapters.DB.Exec("CREATE TABLE IF NOT EXISTS `users` ( `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT, `nickname` varchar(255) NOT NULL, `email` varchar(100) NOT NULL, `user_type` enum('manager','technician') DEFAULT 'technician', `password` varchar(100) NOT NULL, `created_at` datetime DEFAULT CURRENT_TIMESTAMP, `updated_at` datetime DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), UNIQUE KEY `nickname` (`nickname`), UNIQUE KEY `email` (`email`) ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1;")
	if err != nil {
		log.Fatalf("cannot migrated users table: %s", err)
	}
	_, err = adapters.DB.Exec("CREATE TABLE IF NOT EXISTS `tasks` ( `id` bigint(10) unsigned NOT NULL AUTO_INCREMENT, `summary` text NOT NULL, `author_id` bigint(10) unsigned NOT NULL, `date` datetime DEFAULT CURRENT_TIMESTAMP, `created_at` datetime DEFAULT CURRENT_TIMESTAMP, `updated_at` datetime DEFAULT CURRENT_TIMESTAMP, PRIMARY KEY (`id`), KEY `tasks_author_id_users_id_foreign` (`author_id`), CONSTRAINT `tasks_author_id_users_id_foreign` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=latin1;")
	if err != nil {
		log.Fatalf("cannot migrated tasks table")
	}
	log.Printf("Successfully migrated dbs table")
	return nil
}

func RefreshTables() error {
	_, err := adapters.DB.Exec("DELETE FROM `users`;")
	if err != nil {
		log.Fatalf("cannot erase users table: %s", err)
	}
	_, err = adapters.DB.Exec("DELETE FROM `tasks`;")
	if err != nil {
		log.Fatalf("cannot erase tasks table: %s", err)
	}
	log.Printf("Successfully refreshed user table")
	return nil
}

func SeedOneUser() (models.User, error) {
	user := models.User{
		Nickname: "Pet",
		Email:    "pet@gmail.com",
		Password: "password",
		UserType: enums.TECHNICIAN,
	}
	err := user.HashPassword()
	if err != nil {
		return models.User{}, err
	}
	user.Prepare()
	_, err = adapters.DB.Exec("INSERT INTO `users` (`nickname`, `email`, `user_type`, `password`) VALUES (?, ?, ?, ?);", user.Nickname, user.Email, user.UserType, user.Password)
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
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
		users[i].Prepare()
		res, err := adapters.DB.Exec("INSERT INTO `users` (`nickname`, `email`, `user_type`, `password`) VALUES (?, ?, ?, ?);", users[i].Nickname, users[i].Email, users[i].UserType, users[i].Password)
		if err != nil {
			return []models.User{}, err
		}
		lastInsertedId, err := res.LastInsertId()
		if err != nil {
			return []models.User{}, err
		}
		users[i].ID = uint64(lastInsertedId)
	}
	return users, nil
}

func SeedOneUserAndOneTask() (models.User, models.Task, error) {
	user := models.User{
		Nickname: "Sam Phil",
		Email:    "sam@gmail.com",
		Password: "password",
		UserType: enums.TECHNICIAN,
	}
	err := user.HashPassword()
	if err != nil {
		log.Fatalf("cannot seed users table: %v", err)
	}
	user.Prepare()
	res, err := adapters.DB.Exec("INSERT INTO `users` (`nickname`, `email`, `user_type`, `password`) VALUES (?, ?, ?, ?);", user.Nickname, user.Email, user.UserType, user.Password)
	if err != nil {
		return models.User{}, models.Task{}, err
	}
	lastInsertedId, err := res.LastInsertId()
	if err != nil {
		return models.User{}, models.Task{}, err
	}
	task := models.Task{
		ID:       1,
		Summary:  "This is the summary sam",
		AuthorID: uint64(lastInsertedId),
	}
	encryptedTask := task
	err = utils.Encrypt(&encryptedTask.Summary)
	if err != nil {
		log.Fatalf("cannot seed tasks table: %v", err)
	}
	_, err = adapters.DB.Exec("INSERT INTO `tasks` (`summary`, `author_id`) VALUES (?, ?);", encryptedTask.Summary, encryptedTask.AuthorID)
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
			AuthorID: users[2].ID,
		},
		{
			Summary:  "Hello world 2",
			AuthorID: users[3].ID,
		},
	}

	for i := range tasks {
		err = utils.Encrypt(&tasks[i].Summary)
		if err != nil {
			log.Fatalf("cannot seed tasks table: %v", err)
		}
		res, err := adapters.DB.Exec("INSERT INTO `tasks` (`summary`, `author_id`) VALUES (?, ?);", tasks[i].Summary, tasks[i].AuthorID)
		if err != nil {
			log.Fatalf("cannot seed tasks table: %v", err)
		}
		lastInsertedId, err := res.LastInsertId()
		if err != nil {
			return []models.User{}, []models.Task{}, err
		}
		tasks[i].ID = uint64(lastInsertedId)
	}
	return users, tasks, nil
}
