package api

const (
	routeLogin      = "/login"
	routeLogout     = "/logout"
	routeRefresh    = "/refresh"
	routeRegister   = "/register"
	routeUnregister = "/unregister"

	routeDepartments = "/departments"
	routeDepartment  = routeDepartments + "/{id:[0-9]+}"

	routeUsers = "/users"
	routeUser  = routeUsers + "/{id:[0-9]+}"

	routeActivities = "/activities"
	routeActivity   = routeActivities + "/{id:[0-9]+}"

	routeControl             = "/control"
	routeUsersActivity       = routeControl + "/user/{id:[0-9]+}"
	routeDepartmentsActivity = routeControl + "/department/{id:[0-9]+}"
)
