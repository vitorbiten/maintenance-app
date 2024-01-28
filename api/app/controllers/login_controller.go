package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/auth"
	"github.com/vitorbiten/maintenance/api/app/models"
	"golang.org/x/crypto/bcrypt"
)

// Login creates an auth token
//
//	@Summary		Creates an auth token
//	@Tags			login
//	@Produce		json
//	@Param			email		body		models.Email	true	"user email"
//	@Param			password	body		models.Password	true	"user password"
//	@Success		200	{object}	models.User
//	@Failure		422	{object}	nil
//	@Router			/login [post]
func Login(context *gin.Context) {
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
	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	token, err := SignIn(user.Email, user.Password)
	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": "incorrect details"})
		return
	}

	context.JSON(http.StatusOK, token)
}

func SignIn(email, password string) (string, error) {
	user := models.User{}
	_, err := user.FindUserByEmail(adapters.DB, email)
	if err != nil {
		return "", err
	}
	err = models.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", err
	}
	return auth.CreateToken(user.ID)
}
