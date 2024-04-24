package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	md "nfs002/template/v1/internal/models"
	ut "nfs002/template/v1/internal/utils"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"
)

func (app *application) CreateAuthToken(w http.ResponseWriter, r *http.Request) {
	var input md.TokenRequest

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
	ttl := (1 * time.Hour) + (time.Duration(input.Expiry) * time.Minute)
	token, err := md.GenerateToken(user.ID, ttl, input.Scope)
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
		Error   bool      `json:"error"`
		Message string    `json:"message"`
		Token   *md.Token `json:"authentication_token"`
	}

	payload.Error = false
	payload.Message = fmt.Sprintf("token for %s created", input.Email)
	payload.Token = token

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) authenticateToken(r *http.Request, scope []string) (*md.User, *md.Token, error) {
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
	u, ok := r.Context().Value(key).(*md.User)
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

	var user md.User

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
		ut.ErrorLog("Error parsing 'id' parameter", err)
		app.badRequest(w, errors.New("invalid request parameter 'UserID'"))
	}

	var user md.UpdateUserRequest

	if err := app.readJSON(w, r, &user); err != nil {
		app.badRequest(w, err)
		return
	}

	if user.Trim(); user.IsEmpty() {
		app.badRequest(w, errors.New("nothing to update"))
		return
	}

	// Update an existing user
	if err := app.DB.EditUser(userID, user); err != nil {
		ut.ErrorLog("Error updating user", err)
		app.badRequest(w, err)
		return
	}

	if user.Password != "" {
		newHash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
		if err != nil {
			app.internalError(w)
			return
		}

		if err := app.DB.UpdatePasswordForUser(userID, user, string(newHash)); err != nil {
			app.internalError(w)
			return
		}
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "sucesfully updated user"
	app.writeJSON(w, http.StatusOK, resp)
}

// DeleteUser deletes a user, and all associated tokens, from the database
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := strconv.Atoi(id)

	if userID <= 0 || err != nil {
		ut.ErrorLog("Error parsing 'id' parameter", err)
		app.badRequest(w, errors.New("invalid request parameter 'UserID'"))
		return
	}

	if err := app.DB.DeleteUser(userID); err != nil {
		ut.ErrorLog("Error deleting user", err)
		app.badRequest(w, err)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "succesfully deleted user"
	app.writeJSON(w, http.StatusOK, resp)
}
