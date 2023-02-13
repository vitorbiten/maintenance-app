package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/vitorbiten/maintenance/api/app/enums"
	"github.com/vitorbiten/maintenance/api/app/models"
	"gopkg.in/go-playground/assert.v1"
)

func TestCreateUser(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")

	samples := []struct {
		inputJSON    string
		statusCode   int
		nickname     string
		email        string
		errorMessage string
	}{
		{
			inputJSON:    `{"nickname":"New", "email": "new@gmail.com", "password": "password"}`,
			statusCode:   201,
			nickname:     "New",
			email:        "new@gmail.com",
			errorMessage: "",
		},
		{
			inputJSON:    `{"nickname":"Frank", "email": "new@gmail.com", "password": "password"}`,
			statusCode:   500,
			errorMessage: "incorrect details",
		},
		{
			inputJSON:    `{"nickname":"New", "email": "grand@gmail.com", "password": "password"}`,
			statusCode:   500,
			errorMessage: "incorrect details",
		},
		{
			inputJSON:    `{"nickname":"Pet", "email": "pet@gmail.com", "password": "password", "user_type": "manager"}`,
			statusCode:   422,
			errorMessage: "invalid argument",
		},
		{
			inputJSON:    `{"nickname":"Kan", "email": "kangmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "invalid email",
		},
		{
			inputJSON:    `{"nickname": "", "email": "kan@gmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "required nickname",
		},
		{
			inputJSON:    `{"nickname": "Kan", "email": "", "password": "password"}`,
			statusCode:   422,
			errorMessage: "required email",
		},
		{
			inputJSON:    `{"nickname": "Kan", "email": "kan@gmail.com", "password": ""}`,
			statusCode:   422,
			errorMessage: "required password",
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/users", bytes.NewBufferString(v.inputJSON))
		OnError(err, fmt.Sprintf("Error on POST /users: %v", err))
		router.ServeHTTP(rr, req)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 201 {
			assert.Equal(t, responseMap["nickname"], v.nickname)
			assert.Equal(t, responseMap["email"], v.email)
		}
		if v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestGetUsers(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, err := SeedUsers()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	managerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	technicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	samples := []struct {
		tokenGiven   string
		statusCode   int
		errorMessage string
		response     []models.User
	}{
		{
			tokenGiven:   managerTokenString,
			statusCode:   200,
			errorMessage: "",
			response: []models.User{
				{
					ID:       users[2].ID,
					Nickname: "Kenny Morris",
					Email:    "kenny@gmail.com",
					UserType: enums.TECHNICIAN,
				},
				{
					ID:       users[3].ID,
					Nickname: "Denny Morris",
					Email:    "denny@gmail.com",
					UserType: enums.TECHNICIAN,
				},
				{
					ID:       users[4].ID,
					Nickname: "Billy Joe",
					Email:    "billy@gmail.com",
					UserType: enums.TECHNICIAN,
				},
			},
		},
		{
			// When technician token is given
			tokenGiven:   technicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When incorrect token is given
			tokenGiven:   "This is an incorrect token",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When no token is given
			tokenGiven:   "",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/users", nil)
		OnError(err, fmt.Sprintf("Error on GET /users: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			var users []models.User
			err = json.Unmarshal(rr.Body.Bytes(), &users)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, len(users), 3)
			assert.Equal(t, users[0].ID, v.response[0].ID)
			assert.Equal(t, users[1].ID, v.response[1].ID)
			assert.Equal(t, users[2].ID, v.response[2].ID)
		}
		if v.statusCode == 401 && v.errorMessage != "" {
			responseMap := make(map[string]interface{})
			err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestGetUserByID(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, err := SeedUsers()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	secondManagerUser := users[1]
	technicianUser := users[2]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	ManagerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	TechnicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	sample := []struct {
		id           string
		tokenGiven   string
		statusCode   int
		nickname     string
		email        string
		errorMessage string
	}{
		{
			// When manager tries to get themself
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   ManagerTokenString,
			statusCode:   200,
			nickname:     managerUser.Nickname,
			email:        managerUser.Email,
			errorMessage: "",
		},
		{
			// When manager tries to get a technician
			id:           strconv.Itoa(int(technicianUser.ID)),
			tokenGiven:   ManagerTokenString,
			statusCode:   200,
			nickname:     technicianUser.Nickname,
			email:        technicianUser.Email,
			errorMessage: "",
		},
		{
			// When technician tries to get themself
			id:           strconv.Itoa(int(technicianUser.ID)),
			tokenGiven:   TechnicianTokenString,
			statusCode:   200,
			nickname:     technicianUser.Nickname,
			email:        technicianUser.Email,
			errorMessage: "",
		},
		{
			// When incorrect id is given
			id:           strconv.Itoa(999),
			tokenGiven:   ManagerTokenString,
			statusCode:   404,
			errorMessage: "user not found",
		},
		{
			// When manager tries to get another manager user
			id:           strconv.Itoa(int(secondManagerUser.ID)),
			tokenGiven:   ManagerTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When technician tries to get manager user
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   TechnicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When incorrect token is given
			id:           strconv.Itoa(int(technicianUser.ID)),
			tokenGiven:   "This is an incorrect token",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
	}
	for _, v := range sample {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/users/"+v.id, nil)
		OnError(err, fmt.Sprintf("Error on GET /users/id: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			assert.Equal(t, responseMap["nickname"], v.nickname)
			assert.Equal(t, responseMap["email"], v.email)
		}
		if v.statusCode == 401 || v.statusCode == 404 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestUpdateUser(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, err := SeedUsers()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	ManagerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	TechnicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	samples := []struct {
		id             string
		updateJSON     string
		statusCode     int
		updateNickname string
		updateEmail    string
		tokenGiven     string
		errorMessage   string
	}{
		{
			id:             strconv.Itoa(int(managerUser.ID)),
			updateJSON:     `{"nickname":"Grand", "email": "grand@gmail.com", "password": "password"}`,
			statusCode:     200,
			updateNickname: "Grand",
			updateEmail:    "grand@gmail.com",
			tokenGiven:     ManagerTokenString,
			errorMessage:   "",
		},
		{
			id:             strconv.Itoa(int(technicianUser.ID)),
			updateJSON:     `{"nickname":"Tomas", "email": "tomas@gmail.com", "password": "password"}`,
			statusCode:     200,
			updateNickname: "Tomas",
			updateEmail:    "tomas@gmail.com",
			tokenGiven:     TechnicianTokenString,
			errorMessage:   "",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Frank", "email": "tomas@gmail.com", "password": "password"}`,
			statusCode:   500,
			tokenGiven:   ManagerTokenString,
			errorMessage: "incorrect details",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Tomas", "email": "grand@gmail.com", "password": "password"}`,
			statusCode:   500,
			tokenGiven:   ManagerTokenString,
			errorMessage: "incorrect details",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Kan", "email": "kangmail.com", "password": "password"}`,
			statusCode:   422,
			tokenGiven:   ManagerTokenString,
			errorMessage: "invalid email",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname": "", "email": "kan@gmail.com", "password": "password"}`,
			statusCode:   422,
			tokenGiven:   ManagerTokenString,
			errorMessage: "required nickname",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname": "Kan", "email": "", "password": "password"}`,
			statusCode:   422,
			tokenGiven:   ManagerTokenString,
			errorMessage: "required email",
		},
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Woman", "email": "woman@gmail.com", "password": ""}`,
			statusCode:   422,
			tokenGiven:   ManagerTokenString,
			errorMessage: "required password",
		},
		{
			// When technician tries to become a manager
			id:           strconv.Itoa(int(technicianUser.ID)),
			updateJSON:   `{"nickname":"Tomas", "email": "tomas@gmail.com", "password": "password", "user_type": "manager"}`,
			statusCode:   422,
			tokenGiven:   TechnicianTokenString,
			errorMessage: "invalid argument",
		},
		{
			// When manager tries to update a technician
			id:           strconv.Itoa(int(technicianUser.ID)),
			updateJSON:   `{"nickname": "Mike", "email": "mike@gmail.com", "password": "password"}`,
			tokenGiven:   ManagerTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When technician tries to update a manager
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname": "Mike", "email": "mike@gmail.com", "password": "password"}`,
			tokenGiven:   TechnicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When no token was passed
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Man", "email": "man@gmail.com", "password": "password"}`,
			statusCode:   401,
			tokenGiven:   "",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When incorrect token is given
			id:           strconv.Itoa(int(managerUser.ID)),
			updateJSON:   `{"nickname":"Woman", "email": "woman@gmail.com", "password": "password"}`,
			statusCode:   401,
			tokenGiven:   "This is incorrect token",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			id:         "unknwon",
			tokenGiven: ManagerTokenString,
			statusCode: 400,
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", "/users/"+v.id, bytes.NewBufferString(v.updateJSON))
		OnError(err, fmt.Sprintf("Error on PUT /users/id: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			assert.Equal(t, responseMap["nickname"], v.updateNickname)
			assert.Equal(t, responseMap["email"], v.updateEmail)
		}
		if v.statusCode == 401 || v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestDeleteUser(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, err := SeedUsers()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	ManagerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	TechnicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	sample := []struct {
		id           string
		tokenGiven   string
		statusCode   int
		errorMessage string
	}{
		{
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   ManagerTokenString,
			statusCode:   204,
			errorMessage: "",
		},
		{
			// When manager tries to delete a technician
			id:           strconv.Itoa(int(technicianUser.ID)),
			tokenGiven:   ManagerTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When technician tries to update a manager
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   TechnicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When no token is given
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   "",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When incorrect token is given
			id:           strconv.Itoa(int(managerUser.ID)),
			tokenGiven:   "This is an incorrect token",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			id:         "unknwon",
			tokenGiven: ManagerTokenString,
			statusCode: 400,
		},
	}

	for _, v := range sample {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", "/users/"+v.id, nil)
		OnError(err, fmt.Sprintf("Error on DELETE /users/id: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 401 && v.errorMessage != "" {
			responseMap := make(map[string]interface{})
			err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}
