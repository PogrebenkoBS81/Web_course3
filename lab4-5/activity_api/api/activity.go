package api

import (
	"activity_api/api/api_common"
	"activity_api/common/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetActivities - get all activity records
func (a *AApi) GetActivities(w http.ResponseWriter, r *http.Request) {
	activities, err := a.sqlManager.GetActivities()

	entry := a.logger.WithField("func", "GetActivities")
	entry.Debugf("Request from: ", r.RemoteAddr)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetActivities(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with activities list (len %d)", r.RemoteAddr, len(activities))
	api_common.RespondWithJson(w, http.StatusOK, &activities, a.logger)
}

// GetActivity - returns activity record with given ID
func (a *AApi) GetActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "GetActivity")
	entry.Debugf("Request from %s, activityID: %s", r.RemoteAddr, vars["id"])

	activity, err := a.sqlManager.GetActivity(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetActivity(): %v", err),
			a.logger,
		)

		return
	}
	// If activity doesn't exist - return an error.
	if activity == nil {
		entry.Errorf("Respond to %s, department doesn't exist", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusNotFound,
			"depart doesn't exists",
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, *activity)
	api_common.RespondWithJson(w, http.StatusOK, &activity, a.logger)
}

// CreateActivity - creates activity record from given JSON.
func (a *AApi) CreateActivity(w http.ResponseWriter, r *http.Request) {
	entry := a.logger.WithField("func", "CreateActivity")
	entry.Debug("Request from:", r.RemoteAddr)

	activity := new(models.Activity)

	if err := json.NewDecoder(r.Body).Decode(activity); err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("Decode(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Creating activity %+v, request from: %s", activity, r.RemoteAddr)
	id, err := a.sqlManager.CreateActivity(activity)

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("CreateActivity(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Activity created, responding to %s with: %d", r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusCreated, &models.ObjectID{ID: id}, a.logger)
}

// DeleteActivity - deletes activity record with given ID
func (a *AApi) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entry := a.logger.WithField("func", "DeleteActivity")
	entry.Debugf("Request from %s, activityID: %s", r.RemoteAddr, vars["id"])

	id, err := a.sqlManager.DeleteActivity(vars["id"])

	if err != nil {
		entry.Errorf("Respond to %s, error:", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("DeleteActivity(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Activity %s deleted, responding to %s with: %d", vars["id"], r.RemoteAddr, id)
	api_common.RespondWithJson(w, http.StatusOK, &models.ObjectID{ID: id}, a.logger)
}
