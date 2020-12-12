package middleware

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type LoggingMiddleware struct {
	logger logrus.FieldLogger
}

func NewLoggingMiddleware(logger logrus.FieldLogger) *LoggingMiddleware {
	m := new(LoggingMiddleware)
	m.logger = logger.WithField("module", "LoggingMiddleware")

	return m
}

func (l *LoggingMiddleware) LogAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l.logger.WithField("func", "LogAuthMiddleware").
			Debugf("API Logger: %s %s -> %s", r.RemoteAddr, r.Method, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}
