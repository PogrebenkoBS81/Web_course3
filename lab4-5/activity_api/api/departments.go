package api

import (
	"activity_api/api/api_common"
	"activity_api/common/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetDepartments - returns all departments records.
func (a *AApi) GetDepartments(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "GetDepartments")
	entry.Debug("Request from:", r.RemoteAddr)

	departs, err := a.sqlManager.GetDepartments()

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetDepartments(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, &departs)
	api_common.RespondWithJson(w, http.StatusOK, &departs, a.logger)
}

// GetDepartment - returns department record with given ID.
func (a *AApi) GetDepartment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "GetDepartment")
	entry.Debugf("Request from %s, DepartID: %s", r.RemoteAddr, vars["id"])

	depart, err := a.sqlManager.GetDepartment(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetDepartment(): %v", err),
			a.logger,
		)

		return
	}

	if depart == nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(w, http.StatusNotFound, "depart doesn't exists", a.logger)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, depart)
	api_common.RespondWithJson(w, http.StatusOK, &depart, a.logger)
}

// CreateDepartment - writes to db department record from json.
func (a *AApi) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "CreateDepartment")
	entry.Debug("Request from:", r.RemoteAddr)

	depart := new(models.Department)

	if err := json.NewDecoder(r.Body).Decode(depart); err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}

	id, err := a.sqlManager.CreateDepartment(depart)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateDepartment(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Department created, responding to %s with: %d", r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusCreated, &models.ObjectID{ID: id}, a.logger)
}

// DeleteDepartment - deletes department with given ID from DB.
func (a *AApi) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "DeleteDepartment")
	entry.Debugf("Request from %s, userID: %s", r.RemoteAddr, vars["id"])

	id, err := a.sqlManager.DeleteDepartment(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("DeleteDepartment(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Department %s deleted, responding to %s with: %d", vars["id"], r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusOK, &models.ObjectID{ID: id}, a.logger)
}
