package controllers

import (
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/vitorbiten/maintenance/api/app/docs"
)

func InitializeRoutes(r *gin.Engine) {
	// Home Route
	r.GET("/", Home)

	// Login Route
	r.POST("/login", Login)

	//Users routes
	r.POST("/users", CreateUser)
	r.GET("/users", GetUsers)
	r.GET("/users/:id", GetUser)
	r.PUT("/users/:id", UpdateUser)
	r.DELETE("/users/:id", DeleteUser)

	//Tasks routes
	r.POST("/tasks", CreateTask)
	r.GET("/tasks", GetTasks)
	r.GET("/tasks/:id", GetTask)
	r.PUT("/tasks/:id", UpdateTask)
	r.DELETE("/tasks/:id", DeleteTask)

	r.GET("/swagger/*any",
		ginSwagger.WrapHandler(swaggerfiles.Handler))
}
