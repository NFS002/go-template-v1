package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	m "nfs002/template/v1/internal/models"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

// Creates a new user and save it to the DB
func (app *application) NewUser(firstName, lastName, email string) (int, error) {
	customer := m.Customer{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}
	id, err := app.DB.InsertCustomer(customer)
	if err != nil {
		app.errorLog.Println(err)
		return 0, err
	}
	return id, nil
}

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var input m.TokenRequest

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequest(w, err)
		return
	}

	// Assign default values
	input.Defaults()

	// Validate requested scope is within user's scope

	// get the user from the database by email; send error if invalid email
	user, err := app.DB.GetUserByEmail(input.Email)
	if err != nil {
		app.invalidCredentials(w, err)
		return
	}

	// validate the password; send error if invalid password
	validPassword, err := app.passwordMatches(user.Password, input.Password)
	if err != nil {
		app.internalError(w)
		return
	}

	if !validPassword {
		// if passwords not match
		app.invalidCredentials(w, errors.New("incorrect password"))
		return
	}

	// Validate if the user has scope to request the token scope
	if err := user.CanRequestScope(input.Scope); err != nil {
		app.badRequest(w, err)
		return
	}

	// generate the token
	duration := (2 * time.Hour) + (time.Duration(input.Expiry) * time.Minute)
	token, err := m.GenerateToken(user.ID, duration, input.Scope)
	if err != nil {
		app.internalError(w)
		return
	}

	// save to database
	if err := app.DB.InsertToken(token, user); err != nil {
		app.internalError(w)
		return
	}

	// send response
	var payload struct {
		Error   bool     `json:"error"`
		Message string   `json:"message"`
		Token   *m.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", input.Email)
	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) authenticateToken(r *http.Request, scope []string) (*m.User, *m.Token, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return nil, nil, errors.New("no authorization header received")
	}

	headerParts := strings.Split(authorizationHeader, " ")
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return nil, nil, errors.New("no authorization header received")
	}

	tokenStr := headerParts[1]
	if len(tokenStr) != 26 {
		return nil, nil, errors.New("authentication token wrong size")
	}

	// get the user from the tokens table
	user, token, err := app.DB.GetUserForToken(tokenStr)
	if err != nil {
		return nil, nil, errors.New("no matching user found")
	}

	if err := token.HasScope(scope); err != nil {
		return nil, nil, err
	}

	return user, token, nil
}

func (app *application) Hello(w http.ResponseWriter, r *http.Request) {

	// if valid user
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = "Hello!"
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) HelloUser(w http.ResponseWriter, r *http.Request) {
	// validate the token, and get associated user
	key := requestContextKey{Key: "user"}
	u, ok := r.Context().Value(key).(*m.User)
	if !ok {
		app.internalError(w)
		return
	}

	// if valid user
	var payload struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("Hello %s %s (%s)!", u.FirstName, u.LastName, u.Email)
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	allUsers, err := app.DB.GetAllUsers()
	if err != nil {
		app.badRequest(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, allUsers)
}

// GetOneUser gets one user by id (from the url) and returns it as JSON
func (app *application) GetOneUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, _ := strconv.Atoi(id)

	user, err := app.DB.GetOneUser(userID)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, user)
}

func (app *application) CreateUser(w http.ResponseWriter, r *http.Request) {

	var user m.User

	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)

	if err != nil {
		app.badRequest(w, err)
		return
	}

	if err = app.DB.AddUser(user, string(hash)); err != nil {
		app.internalError(w)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "user succesfully created"
	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)

	if userID <= 0 || err != nil {
		app.badRequest(w, errors.New("invalid request parameter 'UserID'"))
	}

	var user m.User

	if err := app.readJSON(w, r, &user); err != nil {
		app.badRequest(w, err)
		return
	}

	// Update an existing user
	if err := app.DB.EditUser(user); err != nil {
		app.badRequest(w, err)
		return
	}

	if user.Password != "" {
		newHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
		if err != nil {
			app.internalError(w)
			return
		}

		if err = app.DB.UpdatePasswordForUser(user, string(newHash)); err != nil {
			app.internalError(w)
			return
		}
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	app.writeJSON(w, http.StatusOK, resp)
}

// DeleteUser deletes a user, and all associated tokens, from the database
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, _ := strconv.Atoi(id)

	err := app.DB.DeleteUser(userID)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	app.writeJSON(w, http.StatusOK, resp)
}
