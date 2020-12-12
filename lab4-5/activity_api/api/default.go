package api

import (
	"activity_api/api/api_common"
	"net/http"
)

func (a *AApi) defHandler(w http.ResponseWriter, r *http.Request) {
	a.logger.WithField("func", "defHandler").
		Debugf("Request from %s on path: %s", r.RemoteAddr, r.RequestURI)
	api_common.RespondWithError(w, http.StatusNotFound, "handler doesn't exist", a.logger)
}
