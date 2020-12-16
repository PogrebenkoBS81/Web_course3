package middleware

import (
	"activity_api/api/api_common"
	"activity_api/api/auth"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
)

type AuthMiddleware struct {
	exclusions map[string]bool
	logger     logrus.FieldLogger
}

func NewAuthMiddleware(logger logrus.FieldLogger, exclusions ...string) *AuthMiddleware {
	m := new(AuthMiddleware)
	m.logger = logger.WithField("module", "AuthMiddleware")
	m.exclusions = make(map[string]bool)

	m.logger.Debugf("Adding exclusions to auth middleware: %v", exclusions)
	for _, path := range exclusions {
		m.exclusions[path] = true
	}

	return m
}

func (m *AuthMiddleware) TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := m.exclusions[mux.CurrentRoute(r).GetName()]; !ok {
			entry := m.logger.WithField("func", "TokenAuthMiddleware")
			entry.Debugf("Request on protected handler %s from %s", r.RequestURI, r.RemoteAddr)
			err := auth.TokenValid(r)

			if err != nil {
				entry.Errorf("Respond to %s, error: %v", r.RemoteAddr, err)
				api_common.RespondWithError(
					w,
					http.StatusUnauthorized,
					fmt.Sprintf("TokenValid(): %v", err),
					m.logger)

				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
