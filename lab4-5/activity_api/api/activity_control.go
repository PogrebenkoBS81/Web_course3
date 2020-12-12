package api

import (
	"activity_api/api/api_common"
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

// GetUsersActivity - returns data about user activity for given period of time
// If no time is set is URL query - all time stat is collected.
func (a *AApi) GetUsersActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timeStart := r.URL.Query().Get("TimeStart")
	timeEnd := r.URL.Query().Get("TimeEnd")

	entry := a.logger.WithField("func", "GetUsersActivity")
	entry.Debugf("Request from %s, url timeStart: %s, timeEnd: %s", r.RemoteAddr, timeStart, timeEnd)

	activity, err := a.sqlManager.GetUserActivity(vars["id"], timeStart, timeEnd)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetUserActivity(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, &activity)
	api_common.RespondWithJson(w, http.StatusOK, &activity, a.logger)
}

func (a *AApi) GetDepartmentsActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	timeStart := r.URL.Query().Get("TimeStart")
	timeEnd := r.URL.Query().Get("TimeEnd")

	entry := a.logger.WithField("func", "GetDepartmentsActivity")
	entry.Debugf("Request from %s, url timeStart: %s, timeEnd: %s", r.RemoteAddr, timeStart, timeEnd)

	activity, err := a.sqlManager.GetDepartmentActivity(vars["id"], timeStart, timeEnd)

	if err != nil {
		entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
		api_common.RespondWithError(
			w,
			http.StatusUnprocessableEntity,
			fmt.Sprintf("GetDepartmentActivity(): %v", err),
			a.logger,
		)

		return
	}

	entry.Debugf("Responding to %s with: %+v", r.RemoteAddr, &activity)
	api_common.RespondWithJson(w, http.StatusOK, &activity, a.logger)
}
