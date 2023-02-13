package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/go-playground/assert.v1"
)

func TestHome(t *testing.T) {
	router := SetupRouter()
	rr := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "/", nil)
	OnError(err, fmt.Sprintf("Error on POST /login: %v", err))
	router.ServeHTTP(rr, req)

	var responseString string
	err = json.Unmarshal(rr.Body.Bytes(), &responseString)
	OnError(err, fmt.Sprintf("Cannot convert to string: %v", err))
	assert.Equal(t, rr.Code, 200)
	assert.Equal(t, responseString, "Welcome to the Maintenance API")

}
