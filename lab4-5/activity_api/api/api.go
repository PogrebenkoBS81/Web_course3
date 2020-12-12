package api

import (
	"activity_api/api/auth"
	"activity_api/api/middleware"
	"activity_api/common/cancellation"
	"activity_api/data_manager/cache"
	"activity_api/data_manager/db/core"
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// AApi - activity api for AAService
type AApi struct {
	router *mux.Router
	server *http.Server
	cancel *cancellation.Token

	auth     auth.IAuth
	token    auth.IToken
	password auth.IPassword

	cacheManager cache.ICacheManager
	sqlManager   core.ISQLDatabase
	logger       logrus.FieldLogger
}

// NewAApi - returns new AApi
func NewAApi(
	addr string,
	sqlManager core.ISQLDatabase,
	cacheManager cache.ICacheManager,
	ctx context.Context,
	logger logrus.FieldLogger,
) *AApi {
	api := &AApi{
		router:       mux.NewRouter(),
		cancel:       cancellation.NewCustomToken(ctx, 1).Close(),
		auth:         auth.NewAuth(cacheManager, logger),
		password:     new(auth.PasswordManager),
		token:        auth.NewToken(logger),
		cacheManager: cacheManager,
		sqlManager:   sqlManager,
		logger:       logger.WithField("module", "AApi"),
	}
	// New http server for api
	api.server = &http.Server{
		Addr:         addr,
		Handler:      api.router,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	// init api routs
	return api.initRoutes()
}

func (a *AApi) initRoutes() *AApi {
	a.logger.WithField("func", "initRoutes").Info("Initializing routes for api...")
	// Init auth middleware
	authMiddleware := middleware.NewAuthMiddleware(
		a.logger,
		routeLogin, // Exclude some routes from authz check
		routeRegister,
		routeRefresh,
	)
	//Init logging middleware
	loggingMiddleware := middleware.NewLoggingMiddleware(a.logger)
	// Add logging middleware and auth middleware to router
	a.router.Use(loggingMiddleware.LogAuthMiddleware, authMiddleware.TokenAuthMiddleware)
	// Default route
	a.router.NotFoundHandler = http.HandlerFunc(a.defHandler)
	// Init authz\auth routes
	a.registerRoute(a.Login, routeLogin, http.MethodPost)
	a.registerRoute(a.Logout, routeLogout, http.MethodPost)
	a.registerRoute(a.Refresh, routeRefresh, http.MethodPost)

	a.registerRoute(a.Register, routeRegister, http.MethodPost)
	a.registerRoute(a.Unregister, routeUnregister, http.MethodDelete)
	// Init department routes
	a.registerRoute(a.CreateDepartment, routeDepartments, http.MethodPost)
	a.registerRoute(a.GetDepartments, routeDepartments, http.MethodGet)
	a.registerRoute(a.GetDepartment, routeDepartment, http.MethodGet)
	a.registerRoute(a.DeleteDepartment, routeDepartment, http.MethodDelete)
	// Init users routes
	a.registerRoute(a.CreateUser, routeUsers, http.MethodPost)
	a.registerRoute(a.GetUsers, routeUsers, http.MethodGet)
	a.registerRoute(a.GetUser, routeUser, http.MethodGet)
	a.registerRoute(a.DeleteUser, routeUser, http.MethodDelete)
	// Init activity routes
	a.registerRoute(a.CreateActivity, routeActivities, http.MethodPost)
	a.registerRoute(a.GetActivities, routeActivities, http.MethodGet)
	a.registerRoute(a.GetActivity, routeActivity, http.MethodGet)
	a.registerRoute(a.DeleteActivity, routeActivity, http.MethodDelete)
	// Init activity check routes
	a.registerRoute(a.GetDepartmentsActivity, routeDepartmentsActivity, http.MethodGet)
	a.registerRoute(a.GetUsersActivity, routeUsersActivity, http.MethodGet)

	return a
}

// registerRoute - route init helper
func (a *AApi) registerRoute(f func(http.ResponseWriter, *http.Request), path string, methods ...string) {
	a.logger.WithField("func", "registerRoute").
		Debugf("Initializing route %s with methods: %v", path, methods)
	a.router.HandleFunc(path, f).Name(path).Methods(methods...) // Name if set for ability to exclude route from authz
}

// Start - starts api server
func (a *AApi) Start() {
	a.logger.WithField("func", "Start").Info("Staring AApi on:", a.server.Addr)
	a.logger.Info("Listen and serve:", a.server.ListenAndServe())
}

// Close - closes api server
func (a *AApi) Close() error {
	a.logger.WithField("func", "Close").Info("Closing AApi....")
	return a.server.Close()
}
