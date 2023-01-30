package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/vitorbiten/maintenance/api/app/models"
)

type Server struct {
	DB     *gorm.DB
	Router *mux.Router
}

func (server *Server) Initialize(DbUser, DbPassword, DbPort, DbHost, DbName string) {
	DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)

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

	log.Println("We are connected to the mysql database")
	server.DB.Debug().AutoMigrate(&models.User{}, &models.Task{})
	server.Router = mux.NewRouter()
	server.initializeRoutes()
}

func (server *Server) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
