package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/vitorbiten/maintenance/api/app/controllers"
	"github.com/vitorbiten/maintenance/api/app/seed"
)

var server = controllers.Server{}

func main() {
	server.Initialize(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"))
	seed.Load(server.DB)

	apiPort := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	log.Println("--------------- Maintenance API ---------------")
	log.Printf("Listening to port %s\n", apiPort)

	server.Run(apiPort)
}
