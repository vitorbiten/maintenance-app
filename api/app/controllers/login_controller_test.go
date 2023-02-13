package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestSignIn(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	user, err := SeedOneUser()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))

	samples := []struct {
		email        string
		password     string
		errorMessage string
	}{
		{
			email:        user.Email,
			password:     "password",
			errorMessage: "",
		},
		{
			email:        "Wrong email",
			password:     "password",
			errorMessage: "user not found",
		},
		{
			email:        user.Email,
			password:     "Wrong password",
			errorMessage: "crypto/bcrypt: hashedPassword is not the hash of the given password",
		},
	}

	for _, v := range samples {
		token, err := SignIn(v.email, v.password)
		if err != nil {
			assert.Equal(t, err, errors.New(v.errorMessage))
		} else {
			assert.NotEqual(t, token, "")
		}
	}
}

func TestLogin(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	_, err = SeedOneUser()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))

	samples := []struct {
		inputJSON    string
		statusCode   int
		email        string
		password     string
		errorMessage string
	}{
		{
			inputJSON:    `{"email": "pet@gmail.com", "password": "password"}`,
			statusCode:   200,
			errorMessage: "",
		},
		{
			inputJSON:    `{"email": "pet@gmail.com", "password": "wrong password"}`,
			statusCode:   422,
			errorMessage: "incorrect details",
		},
		{
			inputJSON:    `{"email": "frank@gmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "incorrect details",
		},
		{
			inputJSON:    `{"email": "kangmail.com", "password": "password"}`,
			statusCode:   422,
			errorMessage: "invalid email",
		},
		{
			inputJSON:    `{"email": "", "password": "password"}`,
			statusCode:   422,
			errorMessage: "required email",
		},
		{
			inputJSON:    `{"email": "kan@gmail.com", "password": ""}`,
			statusCode:   422,
			errorMessage: "required password",
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/login", bytes.NewBufferString(v.inputJSON))
		OnError(err, fmt.Sprintf("Error on POST /login: %v", err))
		router.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			assert.NotEqual(t, rr.Body.String(), "")
		}
		if v.statusCode == 422 && v.errorMessage != "" {
			responseMap := make(map[string]interface{})
			err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}
