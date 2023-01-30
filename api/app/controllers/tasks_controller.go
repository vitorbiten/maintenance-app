package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	rabbitmqAdapter "github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/auth"
	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/responses"
	"github.com/vitorbiten/maintenance/api/app/utils"
)

// CreateTask creates a task
//
//	@Summary		Creates a task
//	@Description	Technicians can: create tasks
//	@Tags			tasks
//	@Produce		json
//	@Param			summary			body		models.Summary	true	"task summary (max length: 2500)"
//	@Param			date			body		models.Date		false	"task date"
//	@Success		200	{object}	models.Task
//	@Failure		401	{object}	nil
//	@Failure		404	{object}	nil
//	@Failure		422	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/tasks [post]
func (server *Server) CreateTask(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	task := models.Task{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	tokenUser, err := user.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if tokenUser.UserType != enums.TECHNICIAN {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = task.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = task.Prepare()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	task.AuthorID = uid
	taskCreated, err := task.SaveTask(server.DB)
	if err != nil {
		formattedError := utils.FormatDBError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	managers, err := user.FindAllManagers(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	var messages []map[string]interface{}
	if len(*managers) > 0 {
		for _, manager := range *managers {
			messages = append(messages, map[string]interface{}{
				"nickname":  tokenUser.Nickname,
				"task_id":   strconv.Itoa(int(task.ID)),
				"task_date": task.Date,
				"email":     manager.Email,
			})
		}
		err = rabbitmqAdapter.PublishMessages(messages, "notification")
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, taskCreated.ID))
	responses.JSON(w, http.StatusCreated, taskCreated)
}

// GetTasks returns all existing tasks
//
//	@Summary		Get tasks
//	@Description	Managers can: get all tasks
//	@Tags			tasks
//	@Produce		json
//	@Success		200	{array}		models.Task
//	@Failure		401	{object}	nil
//	@Failure		404	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/tasks [get]
func (server *Server) GetTasks(w http.ResponseWriter, r *http.Request) {
	user := models.User{}
	task := models.Task{}

	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	tokenUser, err := user.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if tokenUser.UserType == enums.MANAGER {
		allTasks, err := task.FindAllTasks(server.DB)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		responses.JSON(w, http.StatusOK, allTasks)
	} else {
		authorTasks, err := task.FindTasksByAuthorID(server.DB, uint64(tokenUser.ID))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		responses.JSON(w, http.StatusOK, authorTasks)
	}
}

// GetTask returns a task by id
//
//	@Summary		Get task by id
//	@Description	Managers can: get all tasks
//	@Description	Technicians can: get only their tasks
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"task id"
//	@Success		200	{object}	models.Task
//	@Failure		400	{object}	nil
//	@Failure		401	{object}	nil
//	@Failure		404	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/tasks/id [get]
func (server *Server) GetTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := models.User{}

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	task := models.Task{}
	taskReceived, err := task.FindTaskByID(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	requestUser, err := user.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uid != taskReceived.AuthorID && requestUser.UserType != enums.MANAGER {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	responses.JSON(w, http.StatusOK, taskReceived)
}

// UpdateTask updates a task by id
//
//	@Summary		Updates task by id
//	@Description	Technicians can: update only their tasks
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"task id"
//	@Param			summary	body	models.Summary	true	"task summary (max length: 2500)"
//	@Param			date	body	models.Date		false	"task date"
//	@Success		200	{object}	models.Task
//	@Failure		400	{object}	nil
//	@Failure		401	{object}	nil
//	@Failure		422	{object}	nil
//	@Failure		404	{object}	nil
//	@Failure		500	{object}	nil
//	@Router			/tasks/id [put]
func (server *Server) UpdateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := models.User{}

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	task := models.Task{}
	taskReceived, err := task.FindTaskByID(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	uid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	tokenUser, err := user.FindUserByID(server.DB, uint32(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uid != taskReceived.AuthorID || tokenUser.UserType == enums.MANAGER {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = task.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = task.Prepare()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	taskUpdated, err := task.UpdateATask(server.DB, pid)
	if err != nil {
		formattedError := utils.FormatDBError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, taskUpdated)
}

// DeleteTask deletes a task by id
//
//	@Summary		Deletes a task by id
//	@Description	Managers can: delete all tasks
//	@Tags			tasks
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	false	"task id"
//	@Success		204	{object}	nil
//	@Failure		400	{object}	nil
//	@Failure		401	{object}	nil
//	@Failure		404	{object}	nil
//	@Router			/tasks/id [delete]
func (server *Server) DeleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := models.User{}
	task := models.Task{}

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
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
	_, err = task.DeleteATask(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	responses.JSON(w, http.StatusNoContent, "")
}
