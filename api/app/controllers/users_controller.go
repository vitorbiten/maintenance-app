package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/vitorbiten/maintenance/api/app/auth"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/responses"
	"github.com/vitorbiten/maintenance/api/app/utils"
)

// CreateUser creates a user
//
//	@Summary		Creates a user
//	@Description	Technicians can: create users
//	@Tags			users
//	@Produce		json
//	@Param			nickname	body		models.Nickname	true	"user nickname"
//	@Param			email		body		models.Email	true	"user email"
//	@Param			password	body		models.Password	true	"user password"
//	@Success		200	{object}	models.User
//	@Failure		422	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/users [post]
func (server *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = user.Validate("")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user.Prepare()
	userCreated, err := user.SaveUser(server.DB)
	if err != nil {
		formattedError := utils.FormatDBError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.RequestURI, userCreated.ID))
	responses.JSON(w, http.StatusCreated, userCreated)
}

// GetUsers returns all users
//
//	@Summary		Returns all technician users
//	@Description	Managers can: get all technician users
//	@Tags			users
//	@Produce		json
//	@Success		200	{array}		models.User
//	@Failure		400	{object}	nil
//	@Failure		401	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/users [get]
func (server *Server) GetUsers(w http.ResponseWriter, r *http.Request) {
	user := models.User{}

	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	tokenUser, err := user.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if tokenUser.UserType != enums.MANAGER {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	users, err := user.FindAllTechnicians(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, users)
}

// GetUser returns an user
//
//	@Summary		Returns an user
//	@Description	Managers can: get all technician users and themselves
//	@Description	Technicians can: get themselves
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"user id"
//	@Success		200	{object}	models.User
//	@Failure		400 {object}	nil
//	@Failure		401 {object}	nil
//	@Failure		404 {object}	nil
//	@Failure		500	{object}	nil
//	@Router			/users/id [get]
func (server *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	requestedUser := &models.User{}
	requestedUser, err = requestedUser.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	tokenUser := &models.User{}
	tokenUser, err = tokenUser.FindUserByID(server.DB, tokenID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if tokenID != uint32(uid) && tokenUser.UserType != enums.MANAGER {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	if tokenID != uint32(uid) && requestedUser.UserType == enums.MANAGER {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	responses.JSON(w, http.StatusOK, requestedUser)
}

// UpdateUser updates an user
//
//	@Summary		Updates an user by id
//	@Description	Technicians can: update their tasks
//	@Tags			users
//	@Produce		json
//	@Param			id	path		string	true	"user id"
//	@Param			nickname	body		models.Nickname	true	"user nickname"
//	@Param			email		body		models.Email	true	"user email"
//	@Param			password	body		models.Password	true	"user password"
//	@Success		200	{object}	models.User
//	@Failure		400 {object}	nil
//	@Failure		401 {object}	nil
//	@Failure		500	{object}	nil
//	@Router			/users/id [put]
func (server *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["id"], 10, 32)

	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	if tokenID != uint32(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	err = user.Validate("update")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user.Prepare()
	updatedUser, err := user.UpdateAUser(server.DB, uint32(uid))
	if err != nil {
		formattedError := utils.FormatDBError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, updatedUser)
}

// DeleteUser deletes a user by id
//
//	@Summary		Deletes a user by id
//	@Description	Managers can: delete themselves
//	@Description	Technicians can: delete themselves
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	false	"user id"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	nil
//	@Failure		401	{object}	nil
//	@Router			/users/id [delete]
func (server *Server) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := models.User{}

	uid, err := strconv.ParseUint(vars["id"], 10, 32)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	tokenID, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	if tokenID != uint32(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = user.DeleteAUser(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", uid))
	responses.JSON(w, http.StatusNoContent, "")
}
