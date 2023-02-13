package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	rabbitmqAdapter "github.com/vitorbiten/maintenance/api/app/adapters"
	"github.com/vitorbiten/maintenance/api/app/models"
	"github.com/vitorbiten/maintenance/api/app/utils"
	"gopkg.in/go-playground/assert.v1"
)

func TestCreateTask(t *testing.T) {
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
		inputJSON    string
		statusCode   int
		tokenGiven   string
		summary      string
		date         string
		errorMessage string
	}{
		{
			inputJSON:    `{"summary": "the summary"}`,
			statusCode:   201,
			tokenGiven:   technicianTokenString,
			summary:      "the summary",
			errorMessage: "",
		},
		{
			inputJSON:    `{"summary": "the summary", "date": "2011-10-05T14:48:00Z" }`,
			statusCode:   201,
			tokenGiven:   technicianTokenString,
			summary:      "the summary",
			date:         "2011-10-05T14:48:00Z",
			errorMessage: "",
		},
		{
			inputJSON:    `{"summary": ""}`,
			statusCode:   422,
			tokenGiven:   technicianTokenString,
			errorMessage: "required summary",
		},
		{
			inputJSON:    `{"summary": "nano"}`,
			statusCode:   422,
			tokenGiven:   technicianTokenString,
			errorMessage: "summary min length is 5 characters",
		},
		{
			inputJSON:    `{"summary": "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu, consequat vitae, eleifend ac, enim. Aliquam lorem ante, dapibus in, viverra quis, feugiat a, tellus. Phasellus viverra nulla ut metus varius laoreet. Quisque rutrum. Aenean imperdiet. Etiam ultricies nisi vel augue. Curabitur ullamcorper ultricies nisi. Nam eget dui. Etiam rhoncus. Maecenas tempus, tellus eget condimentum rhoncus, sem quam semper libero, sit amet adipiscing sem neque sed ipsum. Nam quam nunc, blandit vel, luctus pulvinar, hendrerit id, lorem. Maecenas nec odio et ante tincidunt tempus. Donec vitae sapien ut libero venenatis faucibus. Nullam quis ante. Etiam sit amet orci eget eros faucibus tincidunt. Duis leo. Sed fringilla mauris sit amet nibh. Donec sodales sagittis magna. Sed consequat, leo eget bibendum sodales, augue velit cursus nunc, quis gravida magna mi a libero. Fusce vulputate eleifend sapien. Vestibulum purus quam, scelerisque ut, mollis sed, nonummy id, metus. Nullam accumsan lorem in dui. Cras ultricies mi eu turpis hendrerit fringilla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia Curae; In ac dui quis mi consectetuer lacinia. Nam pretium turpis et arcu. Duis arcu tortor, suscipit eget, imperdiet nec, imperdiet iaculis, ipsum. Sed aliquam ultrices mauris. Integer ante arcu, accumsan a, consectetuer eget, posuere ut, mauris. Praesent adipiscing. Phasellus ullamcorper ipsum rutrum nunc. Nunc nonummy metus. Vestibulum volutpat pretium libero. Cras id dui. Aenean ut eros et nisl sagittis vestibulum. Nullam nulla eros, ultricies sit amet, nonummy id, imperdiet feugiat, pede. Sed lectus. Donec mollis hendrerit risus. Phasellus nec sem in justo pellentesque facilisis. Etiam imperdiet imperdiet orci. Nunc nec neque. Phasellus leo dolor, tempus non, auctor et, hendrerit quis, nisi. Curabitur ligula sapien, tincidunt non, euismod vitae, posuere imperdiet, leo. Maecenas malesuada. Praesent congue erat at massasi."}`,
			statusCode:   422,
			tokenGiven:   technicianTokenString,
			errorMessage: "summary max length is 2500 characters",
		},
		{
			// When manager token is passed
			inputJSON:    `{"summary": "the summary"}`,
			statusCode:   401,
			tokenGiven:   managerTokenString,
			errorMessage: "unauthorized",
		},
		{
			// When incorrect token is passed
			inputJSON:    `{"summary": "the summary"}`,
			statusCode:   401,
			tokenGiven:   "This is an incorrect token",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When no token is passed
			inputJSON:    `{"summary": "the summary"}`,
			statusCode:   401,
			tokenGiven:   "",
			errorMessage: "token contains an invalid number of segments",
		},
	}
	for _, v := range samples {
		var messagesSent int = 0
		var messageController string

		rabbitmqAdapter.PublishMessages = func(messages []map[string]interface{}, controller string) error {
			messagesSent += len(messages)
			messageController = controller
			return nil
		}

		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/tasks", bytes.NewBufferString(v.inputJSON))
		OnError(err, fmt.Sprintf("Error on POST /tasks: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 201 {
			if v.date == "" {
				assert.Equal(t, responseMap["date"], responseMap["created_at"])
			} else {
				assert.Equal(t, responseMap["date"], v.date)
			}
			assert.Equal(t, responseMap["summary"], v.summary)
			assert.Equal(t, responseMap["author_id"], float64(technicianUser.ID))
			assert.Equal(t, messagesSent, 2)
			assert.Equal(t, messageController, "notification")
		}
		if v.statusCode == 401 || v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestGetTasks(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, _, err := SeedUsersAndTasks()
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
		statusCode   int
		tasksLength  int
		tokenGiven   string
		errorMessage string
	}{
		{
			// When technician token is passed
			statusCode:   200,
			tasksLength:  1,
			tokenGiven:   technicianTokenString,
			errorMessage: "",
		},
		{
			// When manager token is passed
			statusCode:   200,
			tasksLength:  2,
			tokenGiven:   managerTokenString,
			errorMessage: "",
		},
		{
			// When incorrect token is passed
			statusCode:   401,
			tokenGiven:   "This is an incorrect token",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When no token is passed
			statusCode:   401,
			tokenGiven:   "",
			errorMessage: "token contains an invalid number of segments",
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/tasks", nil)
		OnError(err, fmt.Sprintf("Error on GET /tasks: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			var tasks []models.Task
			err = json.Unmarshal(rr.Body.Bytes(), &tasks)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, rr.Code, http.StatusOK)
			assert.Equal(t, len(tasks), v.tasksLength)
		}
		if v.statusCode == 401 || v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			responseMap := make(map[string]interface{})
			err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
			OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestFindTaskByID(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, tasks, err := SeedUsersAndTasks()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	firstTechnicianTask := tasks[0]
	err = utils.Decrypt(&firstTechnicianTask.Summary)
	OnError(err, fmt.Sprintf("Cannot decrypt summary: %v", err))
	secondTechnicianTask := tasks[1]
	err = utils.Decrypt(&secondTechnicianTask.Summary)
	OnError(err, fmt.Sprintf("Cannot decrypt summary: %v", err))
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	managerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	technicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	samples := []struct {
		id           string
		statusCode   int
		tokenGiven   string
		summary      string
		author_id    uint64
		errorMessage string
	}{
		{
			id:         strconv.Itoa(int(firstTechnicianTask.ID)),
			statusCode: 200,
			tokenGiven: managerTokenString,
			summary:    firstTechnicianTask.Summary,
			author_id:  firstTechnicianTask.AuthorID,
		},
		{
			id:         strconv.Itoa(int(secondTechnicianTask.ID)),
			statusCode: 200,
			tokenGiven: managerTokenString,
			summary:    secondTechnicianTask.Summary,
			author_id:  secondTechnicianTask.AuthorID,
		},
		{
			id:         strconv.Itoa(int(firstTechnicianTask.ID)),
			statusCode: 200,
			tokenGiven: technicianTokenString,
			summary:    firstTechnicianTask.Summary,
			author_id:  firstTechnicianTask.AuthorID,
		},
		{
			id:           strconv.Itoa(int(secondTechnicianTask.ID)),
			statusCode:   401,
			tokenGiven:   technicianTokenString,
			errorMessage: "unauthorized",
		},
		{
			// When incorrect id is passed
			id:           "999",
			statusCode:   404,
			tokenGiven:   managerTokenString,
			errorMessage: "task not found",
		},
		{
			// When incorrect token is passed
			id:           strconv.Itoa(int(secondTechnicianTask.ID)),
			statusCode:   401,
			tokenGiven:   "This is an incorrect token",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When no token is passed
			id:           strconv.Itoa(int(secondTechnicianTask.ID)),
			statusCode:   401,
			tokenGiven:   "",
			errorMessage: "token contains an invalid number of segments",
		},
		{
			id:         "unknwon",
			statusCode: 400,
		},
	}
	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("GET", "/tasks/"+v.id, nil)
		OnError(err, fmt.Sprintf("Error on GET /tasks/id: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		assert.Equal(t, rr.Code, v.statusCode)
		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to jsosn: %v", err))

		if v.statusCode == 200 {
			assert.Equal(t, responseMap["summary"], v.summary)
			assert.Equal(t, responseMap["author_id"], float64(v.author_id))
		}
		if v.statusCode == 401 || v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestUpdateTask(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, tasks, err := SeedUsersAndTasks()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	firstTechnicianTask := tasks[0]
	secondTechnicianTask := tasks[1]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	managerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	technicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	samples := []struct {
		id           string
		updateJSON   string
		summary      string
		author_id    int
		tokenGiven   string
		statusCode   int
		errorMessage string
	}{
		{
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			updateJSON:   `{"summary": "This is the updated summary"}`,
			summary:      "This is the updated summary",
			author_id:    int(technicianUser.ID),
			tokenGiven:   technicianTokenString,
			statusCode:   200,
			errorMessage: "",
		},
		{
			id:           strconv.Itoa(int(secondTechnicianTask.ID)),
			updateJSON:   `{"summary": "This is the updated summary"}`,
			tokenGiven:   technicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			updateJSON:   `{"summary": ""}`,
			tokenGiven:   technicianTokenString,
			statusCode:   422,
			errorMessage: "required summary",
		},
		{
			id:           strconv.Itoa(int(999)),
			updateJSON:   `{"summary": ""}`,
			tokenGiven:   technicianTokenString,
			statusCode:   500,
			errorMessage: "task not found",
		},
		{
			// When manager token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			updateJSON:   `{"summary": "This is the updated summary"}`,
			tokenGiven:   managerTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When no token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			updateJSON:   `{"summary": "This is the updated summary"}`,
			tokenGiven:   "",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When incorrect token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			updateJSON:   `{"summary": "This is the updated summary"}`,
			tokenGiven:   "this is an incorrect token",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			id:         "unknwon",
			statusCode: 400,
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("PUT", "/tasks/"+v.id, bytes.NewBufferString(v.updateJSON))
		OnError(err, fmt.Sprintf("Error on PUT /tasks/id: %v", err))
		req.Header.Set("Authorization", v.tokenGiven)
		router.ServeHTTP(rr, req)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(rr.Body.Bytes(), &responseMap)
		OnError(err, fmt.Sprintf("Cannot convert to json: %v", err))
		assert.Equal(t, rr.Code, v.statusCode)
		if v.statusCode == 200 {
			assert.Equal(t, responseMap["summary"], v.summary)
			assert.Equal(t, responseMap["author_id"], float64(v.author_id))
		}
		if v.statusCode == 401 || v.statusCode == 422 || v.statusCode == 500 && v.errorMessage != "" {
			assert.Equal(t, responseMap["error"], v.errorMessage)
		}
	}
}

func TestDeleteTask(t *testing.T) {
	err := RefreshTables()
	OnError(err, "Error refreshing users table")
	users, tasks, err := SeedUsersAndTasks()
	OnError(err, fmt.Sprintf("Error seeding user: %v\n", err))
	managerUser := users[0]
	technicianUser := users[2]
	firstTechnicianTask := tasks[0]
	secondTechnicianTask := tasks[1]
	managerToken, err := SignIn(managerUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as manager: %v\n", err))
	managerTokenString := fmt.Sprintf("Bearer %v", managerToken)
	technicianToken, err := SignIn(technicianUser.Email, "password")
	OnError(err, fmt.Sprintf("Cannot login as technician: %v\n", err))
	technicianTokenString := fmt.Sprintf("Bearer %v", technicianToken)

	samples := []struct {
		id           string
		tokenGiven   string
		statusCode   int
		errorMessage string
	}{
		{
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			tokenGiven:   managerTokenString,
			statusCode:   204,
			errorMessage: "",
		},
		{
			id:           strconv.Itoa(int(secondTechnicianTask.ID)),
			tokenGiven:   managerTokenString,
			statusCode:   204,
			errorMessage: "",
		},
		{
			id:           strconv.Itoa(999),
			tokenGiven:   managerTokenString,
			statusCode:   404,
			errorMessage: "",
		},
		{
			// When technician token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			tokenGiven:   technicianTokenString,
			statusCode:   401,
			errorMessage: "unauthorized",
		},
		{
			// When no token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			tokenGiven:   "",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			// When incorrect token is provided
			id:           strconv.Itoa(int(firstTechnicianTask.ID)),
			tokenGiven:   "this is an incorrect token",
			statusCode:   401,
			errorMessage: "token contains an invalid number of segments",
		},
		{
			id:         "unknwon",
			statusCode: 400,
		},
	}

	for _, v := range samples {
		router := SetupRouter()
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("DELETE", "/tasks/"+v.id, nil)
		OnError(err, fmt.Sprintf("Error on DELETE /tasks/id: %v", err))
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
