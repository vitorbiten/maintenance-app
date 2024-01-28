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
func CreateTask(context *gin.Context) {
	user := models.User{}
	task := models.Task{}

	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if tokenUser.UserType != enums.TECHNICIAN {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = task.Validate()
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = task.Prepare()
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	task.AuthorID = uid
	taskCreated, err := task.SaveTask(tx)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	task.ID = uint64(taskCreated)
	managers, err := user.FindAllManagers(tx)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var messages []map[string]interface{}
	if len(*managers) > 0 {
		for _, manager := range *managers {
			messages = append(messages, map[string]interface{}{
				"nickname":  tokenUser.Nickname,
				"task_id":   strconv.Itoa(int(taskCreated)),
				"task_date": task.Date,
				"email":     manager.Email,
			})
		}
		err = adapters.PublishMessages(messages, "notification")
		if err != nil {
			context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusCreated, task)
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
func GetTasks(context *gin.Context) {
	user := models.User{}
	task := models.Task{}

	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if tokenUser.UserType == enums.MANAGER {
		allTasks, err := task.FindAllTasks(adapters.DB)
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
		context.JSON(http.StatusOK, allTasks)
	} else {
		authorTasks, err := task.FindTasksByAuthorID(adapters.DB, uint64(tokenUser.ID))
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
		context.JSON(http.StatusOK, authorTasks)
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
func GetTask(context *gin.Context) {
	user := models.User{}
	task := models.Task{}

	pid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	requestUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	taskReceived, err := task.FindTaskByID(adapters.DB, pid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	if uid != taskReceived.AuthorID && requestUser.UserType != enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, taskReceived)
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
func UpdateTask(context *gin.Context) {
	user := models.User{}
	task := models.Task{}

	tid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	taskReceived, err := task.FindTaskByID(adapters.DB, tid)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if uid != taskReceived.AuthorID || tokenUser.UserType == enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	body, err := io.ReadAll(context.Request.Body)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = json.Unmarshal(body, &task)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = task.Validate()
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	err = task.Prepare()
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	taskUpdated, err := task.UpdateATask(adapters.DB, tid)
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
	context.JSON(http.StatusOK, taskUpdated)
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
func DeleteTask(context *gin.Context) {
	user := models.User{}
	task := models.Task{}

	pid, err := strconv.ParseUint(context.Param("id"), 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uid, err := auth.ExtractTokenID(context.Request)
	if err != nil {
		context.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	tx, err := adapters.DB.Begin()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tokenUser, err := user.FindUserByID(tx, uid)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if tokenUser.UserType != enums.MANAGER {
		context.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	res, err := task.DeleteATask(adapters.DB, pid)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if res == 0 {
		context.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		context.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	context.Header("Entity", fmt.Sprintf("%d", pid))
	context.JSON(http.StatusNoContent, "")
}
