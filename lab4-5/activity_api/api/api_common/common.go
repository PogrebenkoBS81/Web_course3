package api_common

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net/http"
)

// RespondWithError - responds with error message to client.
func RespondWithError(w http.ResponseWriter, code int, message string, logger logrus.FieldLogger) {
	RespondWithJson(w, code, map[string]string{"Error": message}, logger)
}

// RespondWithJson - responds to client with given data and code.
func RespondWithJson(w http.ResponseWriter, code int, payload interface{}, logger logrus.FieldLogger) {
	entry := logger.WithField("func", "RespondWithJson")
	response, err := json.Marshal(payload)

	if err != nil {
		entry.Warn("Marshall error:", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if _, err := w.Write(response); err != nil {
		entry.Warn("Response write error:", err)
	}

	return
}
