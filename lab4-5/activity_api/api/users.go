package api

import (
	"activity_api/api/api_common"
	"activity_api/common/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetUsers - returns all users. If departmentID was specified in URL query - return all users by department
func (a *AApi) GetUsers(w http.ResponseWriter, r *http.Request) {
	depID := r.URL.Query().Get("departmentID")

	entry := a.logger.WithField("func", "GetUsers")
	entry.Debugf("Request from %s, url departmentID: %s", r.RemoteAddr, depID)

	users, err := a.sqlManager.GetUsers(depID)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetUsers(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, &users)
	api_common.RespondWithJson(w, http.StatusOK, &users, a.logger)
}

// GetUser - returns user with given ID
func (a *AApi) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "GetUser")
	entry.Debugf("Request from %s, userID: %s", r.RemoteAddr, vars["id"])

	user, err := a.sqlManager.GetUser(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetUser(): %v", err),
			a.logger,
		)

		return
	}
	// Id user is nil == doesn't exists
	if user == nil {
		entry.Warnf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(w, http.StatusNotFound, "user doesn't exists", a.logger)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, user)
	api_common.RespondWithJson(w, http.StatusOK, &user, a.logger)
}

// CreateUser - writes user to DB from given JSON.
func (a *AApi) CreateUser(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "CreateUser")
	entry.Debug("Request from:", r.RemoteAddr)

	user := new(models.User)

	if err := json.NewDecoder(r.Body).Decode(user); err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Creating user %+v, request from: %s", user, r.RemoteAddr)
	id, err := a.sqlManager.CreateUser(user)

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateUser(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("User created, responding to %s with: %d", r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusCreated, &models.ObjectID{ID: id}, a.logger)
}

// DeleteUser - deletes user with given ID from DB.
func (a *AApi) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "DeleteUser")
	entry.Debugf("Request from %s, userID: %s", r.RemoteAddr, vars["id"])

	id, err := a.sqlManager.DeleteUser(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("DeleteUser(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("User %s deleted, responding to %s with: %d", vars["id"], r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusOK, &models.ObjectID{ID: id}, a.logger)
}
