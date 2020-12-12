package core

import "activity_api/common/models"

// TODO: Add modify(update) queries
// TODO: Segregate interface into something like: UserManager, DepartmentManager, etc
// ISQLDatabase - database interface for AAService
type ISQLDatabase interface {
	ISQLCore

	CreateDB() error
	Describe() string

	CreateAdmin(admin *models.Admin) (int64, error)
	GetAdmin(name string) (*models.Admin, error)
	DeleteAdmin(name string) (int64, error)

	CreateDepartment(depart *models.Department) (int64, error)
	GetDepartments() ([]*models.Department, error)
	GetDepartment(departID string) (*models.Department, error)
	DeleteDepartment(departID string) (int64, error)

	CreateUser(user *models.User) (int64, error)
	GetUsers(depID string) ([]*models.User, error)
	GetUser(userID string) (*models.User, error)
	DeleteUser(userID string) (int64, error)

	CreateActivity(activity *models.Activity) (int64, error)
	GetActivities() ([]*models.Activity, error)
	GetActivity(activityID string) (*models.Activity, error)
	DeleteActivity(activityID string) (int64, error)

	GetUserActivity(userID, timeBefore, timeAfter string) (*models.UserActivity, error)
	GetDepartmentActivity(departID, timeBefore, timeAfter string) (*models.DepartmentActivity, error)
}
