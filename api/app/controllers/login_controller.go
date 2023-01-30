package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/vitorbiten/maintenance/api/app/auth"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/responses"
	"github.com/vitorbiten/maintenance/api/app/utils"
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
func (server *Server) Login(w http.ResponseWriter, r *http.Request) {
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
	user.Prepare()
	err = user.Validate("login")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	token, err := server.SignIn(user.Email, user.Password)
	if err != nil {
		formattedError := utils.FormatDBError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, token)
}

func (server *Server) SignIn(email, password string) (string, error) {
	user := models.User{}

	err := server.DB.Debug().Model(models.User{}).Where("email = ?", email).Take(&user).Error
	if err != nil {
		return "", err
	}
	err = models.VerifyPassword(user.Password, password)
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword {
		return "", err
	}
	return auth.CreateToken(user.ID)
}
