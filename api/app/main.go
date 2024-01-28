package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/joho/godotenv/autoload"
	"github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/controllers"
	"golang.org/x/sync/errgroup"

	"context"
	"net/http"
	"os/signal"
	"syscall"
)

func serveApplication() {
	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:8000"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "hx-target", "hx-current-url", "hx-request"}

	router.Use(cors.New(config))

	controllers.InitializeRoutes(router)

	apiPort := fmt.Sprintf(":%s", os.Getenv("API_PORT"))
	server := &http.Server{
		Addr:    apiPort,
		Handler: router,
	}

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		log.Printf("Listening to port %s\n", apiPort)
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		log.Println("Shutting down server...")
		return server.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Printf("Exit reason: %s \n", err)
	}
}

func main() {
	adapters.LoadDatabase()
	serveApplication()
}
