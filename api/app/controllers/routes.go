package controllers

import (
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/vitorbiten/maintenance/api/app/docs"
	"github.com/vitorbiten/maintenance/api/app/middlewares"
)

func (s *Server) initializeRoutes() {
	docs.SwaggerInfo.Title = "Maintenance API Docs"
	docs.SwaggerInfo.Description = "This is a sample maintenance tasks server."
	docs.SwaggerInfo.Version = "1.0"

	// Home Route
	s.Router.HandleFunc("/", middlewares.SetMiddlewareJSON(s.Home)).Methods("GET")

	// Swagger Route
	s.Router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	// Login Route
	s.Router.HandleFunc("/login", middlewares.SetMiddlewareJSON(s.Login)).Methods("POST")

	//Users routes
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(s.CreateUser)).Methods("POST")
	s.Router.HandleFunc("/users", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.GetUsers))).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.GetUser))).Methods("GET")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateUser))).Methods("PUT")
	s.Router.HandleFunc("/users/{id}", middlewares.SetMiddlewareAuthentication(middlewares.SetMiddlewareAuthentication(s.DeleteUser))).Methods("DELETE")

	//Tasks routes
	s.Router.HandleFunc("/tasks", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.CreateTask))).Methods("POST")
	s.Router.HandleFunc("/tasks", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.GetTasks))).Methods("GET")
	s.Router.HandleFunc("/tasks/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.GetTask))).Methods("GET")
	s.Router.HandleFunc("/tasks/{id}", middlewares.SetMiddlewareJSON(middlewares.SetMiddlewareAuthentication(s.UpdateTask))).Methods("PUT")
	s.Router.HandleFunc("/tasks/{id}", middlewares.SetMiddlewareAuthentication(middlewares.SetMiddlewareAuthentication(s.DeleteTask))).Methods("DELETE")
}
