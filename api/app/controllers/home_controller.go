package controllers

import (
	"net/http"

	"github.com/vitorbiten/maintenance/api/app/responses"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Welcome to the Maintenance API")
}
