package sqlite

const (
	createAdminsTable = `
CREATE TABLE IF NOT EXISTS admins (
    admin_name TEXT NOT NULL PRIMARY KEY,
    password_hash TEXT NOT NULL
);`

	createDepartmentsTable = `
CREATE TABLE IF NOT EXISTS department_list (
	department_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	department_name TEXT NOT NULL
);`

	createUsersTable = `
CREATE TABLE IF NOT EXISTS user_list (
	user_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_name TEXT NOT NULL,
	department_id INTEGER NOT NULL,
	CONSTRAINT user_list_FK FOREIGN KEY (department_id) REFERENCES department_list(department_id)
);`

	createActivityTable = `
CREATE TABLE IF NOT EXISTS user_activity (
	record_id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	total_time INTEGER NOT NULL,
	active_time INTEGER NOT NULL,
	activity_date INTEGER NOT NULL,
	CONSTRAINT user_activity_FK FOREIGN KEY (user_id) REFERENCES user_list(user_id)
);`

	adminFind = `
SELECT admin_name
    , password_hash
FROM admins 
WHERE admin_name = ?;`

	adminCreate = `
INSERT INTO admins (admin_name, password_hash)
VALUES(?, ?);`

	adminDelete = `
DELETE FROM admins 
WHERE admin_name = ?;`

	departmentsGet = `
SELECT department_id, department_name 
FROM department_list`

	departmentGet = departmentsGet + `
WHERE department_id = ?;`

	departmentCreate = `
INSERT INTO department_list (department_name)
VALUES (?);`

	departmentDelete = `
DELETE FROM department_list 
WHERE department_id = ?;`

	userCreate = `
INSERT INTO user_list (user_name, department_id)
VALUES (?, ?);`

	usersGet = `
SELECT user_id
    , user_name 
    , department_id 
FROM user_list`

	userGet = usersGet + `
WHERE user_id = ?;`

	userDepartmentGet = usersGet + `
WHERE department_id = ?;`

	userDelete = `
DELETE FROM user_list 
WHERE user_id = ?;`

	activityCreate = `
INSERT INTO user_activity (
	user_id
    , active_time
    , total_time
    , activity_date
)
VALUES (?, ?, ?, ?);`

	activitiesGet = `
SELECT user_activity 
    record_id 
    , user_id
    , active_time
    , total_time
    , activity_date
FROM user_activity`

	activityGet = activitiesGet + `
WHERE record_id = ?;`

	activityDelete = `
DELETE FROM user_activity 
WHERE record_id = ?;`

	getUsersActivity = `
SELECT ua.user_id AS user_id 
	, SUM(ua.total_time) AS total_time 
    , SUM(ua.active_time) AS active_time
FROM user_activity ua
WHERE ua.user_id = ?`

	getDepartmentsActivity = `
SELECT dl.department_id AS user_id 
    , SUM(ua.total_time) AS total_time 
    , SUM(ua.active_time) AS active_time
FROM user_activity ua
INNER JOIN user_list ul 
ON ua.user_id = ul.user_id 
INNER JOIN department_list dl 
ON ul.department_id = dl.department_id 
WHERE dl.department_id = ?`

	activityTimeStart = `
AND ua.activity_date > '%s'`

	activityTimeEnd = `
AND ua.activity_date < '%s'`
)
