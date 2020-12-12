package models

// Admin - admin user of AAService.
type Admin struct {
	Username string `db:"admin_name"`
	// Password - field used only in login and register.
	Password string
	// Hash - password hash with salt.
	Hash string `db:"password_hash"`
}

// Department - AAService Department.
type Department struct {
	DepartmentID   int64  `db:"department_id"`
	DepartmentName string `db:"department_name"`
}

// User - AAService User.
type User struct {
	UserID   int64  `db:"user_id"`
	UserName string `db:"user_name"`
	// DepartmentID - department in which the user participates.
	DepartmentID int64 `db:"department_id"`
}

// Activity - AAService activity.
type Activity struct {
	// UserID - id of user which activity were recorded.
	UserID   int64 `db:"user_id"`
	RecordID int64 `db:"record_id"`
	// TotalTime - total user work time.
	TotalTime int64 `db:"total_time"`
	// ActiveTime - active user time from all total time.
	ActiveTime int64 `db:"active_time"`
	// Date - time when activity record were taken.
	Date int64 `db:"date"`
}

// UserActivity - data about user activity
// Due to rows could be null (which would lead to an err during unmarshalling), pointer are used
type UserActivity struct {
	UserID     *int64 `db:"user_id"`
	ActiveTime *int64 `db:"active_time"`
	TotalTime  *int64 `db:"total_time"`
}

// DepartmentActivity - data about department activity
// Due to rows could be null (which would lead to an err during unmarshalling), pointer are used
type DepartmentActivity struct {
	DepartmentID *int64 `db:"department_id"`
	ActiveTime   *int64 `db:"active_time"`
	TotalTime    *int64 `db:"total_time"`
}

// ObjectID - used when returning last inserted id or rows affected.
type ObjectID struct {
	ID int64
}
