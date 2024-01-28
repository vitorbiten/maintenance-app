package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/auth"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
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
func CreateUser(context *gin.Context) {
	body, err := io.ReadAll(context.Request.Body)

	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = user.Validate("")
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	user.Prepare()
	userCreated, err := user.SaveUser(adapters.DB)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "incorrect details"})
		return
	}
	context.JSON(http.StatusCreated, userCreated)
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
func GetUsers(context *gin.Context) {
	user := models.User{}

	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	tokenUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if tokenUser.UserType != enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	users, err := user.FindAllTechnicians(adapters.DB)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, users)
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
func GetUser(context *gin.Context) {
	uid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenID, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		fmt.Println(err)
		return
	}
	requestedUser := &models.User{}
	requestedUser, err = requestedUser.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	tokenUser := &models.User{}
	tokenUser, err = tokenUser.FindUserByID(tx, tokenID)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if tokenID != uid && tokenUser.UserType != enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if tokenID != uid && requestedUser.UserType == enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, requestedUser)
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
func UpdateUser(context *gin.Context) {
	uid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	user := models.User{}
	err = json.Unmarshal(body, &user)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	tokenID, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if tokenID != uid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err = user.Validate("update")
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	user.Prepare()
	updatedUser, err := user.UpdateAUser(adapters.DB, uid)
	if err != nil {
		if err.Error() == "user not found" {
			context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		context.JSON(http.StatusInternalServerError, gin.H{"error": "incorrect details"})
		return
	}
	context.JSON(http.StatusOK, updatedUser)
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
func DeleteUser(context *gin.Context) {
	user := models.User{}

	uid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokenID, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if tokenID != uid {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	res, err := user.DeleteAUser(adapters.DB, uid)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if res == 0 {
		context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	context.Header("Entity", fmt.Sprintf("%d", uid))
	context.JSON(http.StatusNoContent, "")
}
