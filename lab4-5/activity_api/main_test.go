// Global functional test.
// This test is quite messy, so it would be better to rewrite them in future.
// TODO: Clean up this test,
// TODO: implement unit tests, with DB mocking, etc

package main

import (
	"activity_api/common/http_client"
	"activity_api/common/models"
	"activity_api/control"
	"activity_api/data_manager/cache"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

// Predefined config for AAService
var config = control.AAServiceConfig{
	DbType:     0,
	ConnString: "functional_test_db.db",
	Addr:       "localhost:9332",
	LogLevel:   4,                     // Info level
	LogFile:    "functional_test.log", // temporary unused
	Cache: &cache.ICacheConfig{
		Address:  "localhost:6379",
		Password: "",
		DB:       0,
	},
}

// getTestAdmin - returns randomly created admin.
func getTestAdmin(t *testing.T) []byte {
	log.Println("TEST: TEST: getting test admin")

	u := models.Admin{
		Username: uuid.New().String(),
		// I didn't truly create hash here, because it's only a test,
		// but if sometime client would be implemented - has should be calculated on the client side.
		Hash: uuid.New().String(),
	}

	bts, err := json.Marshal(u)

	if err != nil {
		t.Fatal(err)
	}

	return bts
}

// login - attempts to login with given admin.
func login(t *testing.T, client *http_client.NetHTTP, user []byte) (map[string]string, error) {
	log.Println("TEST: log in API")

	bts, err := client.MakeRequest(http.MethodPost, "http://localhost:9332/login", nil, user)

	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	if err = json.Unmarshal(bts, &m); err != nil {
		t.Fatal(err)
	}

	return m, nil
}

// postObject - posts given data on given handler.
func postObject(t *testing.T, client *http_client.NetHTTP, headers map[string]string, path string, data interface{}) int64 {
	bts, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}

	bts, err = client.MakeRequest(http.MethodPost, path, headers, bts)
	if err != nil {
		t.Fatal(err)
	}

	id := new(models.ObjectID)

	if err = json.Unmarshal(bts, &id); err != nil {
		t.Fatal(err)
	}

	return id.ID
}

// registerAndLogin - register given admin and logins it, returns auth token
func registerAndLogin(t *testing.T, client *http_client.NetHTTP, userBts []byte) map[string]string {
	log.Println("TEST: Registering and login")

	_, err := client.MakeRequest(http.MethodPost, "http://localhost:9332/register", nil, userBts)

	if err != nil {
		t.Fatal(err)
	}

	token, err := login(t, client, userBts)
	if err != nil {
		t.Fatal(err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token["access_token"],
	}

	return headers
}

// unregister - unregister given admin.
func unregister(t *testing.T, client *http_client.NetHTTP, headers map[string]string, userBts []byte) {
	log.Println("TEST: Unregistering")
	_, err := client.MakeRequest(http.MethodDelete, "http://localhost:9332/unregister", headers, userBts)

	if err != nil {
		t.Fatal(err)
	}
}

// addDepartments - adds randomly generated departments.
func addDepartments(t *testing.T, client *http_client.NetHTTP, headers map[string]string) []models.Department {
	log.Println("TEST: Adding random departments")
	deps := make([]models.Department, 0)

	for i := 0; i < rand.Intn(100); i++ {
		dep := models.Department{
			DepartmentName: uuid.New().String(),
		}

		id := postObject(t, client, headers, "http://localhost:9332/departments", dep)
		dep.DepartmentID = id
		deps = append(deps, dep)
	}

	return deps
}

// addUsers - adds randomly generated users to given departments.
func addUsers(
	t *testing.T,
	client *http_client.NetHTTP,
	headers map[string]string,
	deps []models.Department,
) []models.User {
	log.Println("TEST: Adding random users")
	users := make([]models.User, 0)

	for _, dep := range deps {
		for i := 0; i < rand.Intn(10); i++ {
			user := models.User{
				UserName:     uuid.New().String(),
				DepartmentID: dep.DepartmentID,
			}

			id := postObject(t, client, headers, "http://localhost:9332/users", user)

			user.UserID = id
			users = append(users, user)
		}
	}

	return users
}

// addActivities - adds randomly generated activities to given users.
func addActivities(
	t *testing.T,
	client *http_client.NetHTTP,
	headers map[string]string,
	users []models.User,
) []models.Activity {
	log.Println("TEST: Adding random activities")

	ids := make([]models.Activity, 0)
	var min int64 = 500
	var max int64 = 1000

	for _, u := range users {
		for i := 0; i < rand.Intn(100); i++ {
			user := models.Activity{
				UserID:     u.UserID,
				TotalTime:  rand.Int63n(max-min) + min,
				ActiveTime: rand.Int63n(min),
				Date:       time.Now().Unix() + rand.Int63n(max),
			}

			id := postObject(t, client, headers, "http://localhost:9332/activities", user)

			user.RecordID = id
			ids = append(ids, user)
		}
	}

	return ids
}

// manualActivityCalc - manually calculates activity time for given users to check api.
func manualActivityCalc(usersID map[int64]bool, minTime, maxTime int64, activity []models.Activity) (int64, int64) {
	log.Println("TEST: Manually calculating activity")

	var manualActive int64
	var manualTotal int64

	for _, a := range activity {
		if _, ok := usersID[a.UserID]; !ok {
			continue
		}

		if a.Date > minTime && a.Date < maxTime {
			manualActive += a.ActiveTime
			manualTotal += a.TotalTime
		}
	}

	return manualActive, manualTotal
}

// getUserTiming - get timing between first and last record, to check URL time query.
func getUserTiming(userID int64, activity []models.Activity) (int64, int64) {
	log.Println("TEST: Creating data time slice")

	sliceMinTime := activity[0].Date + 1 // Don't include first user record to test URL time query
	sliceMaxTime := activity[0].Date

	for _, a := range activity {
		if a.UserID == userID && sliceMaxTime < a.Date {
			sliceMaxTime = a.Date - 1 // Don't include last user record to test URL time query
		}
	}

	return sliceMinTime, sliceMaxTime
}

// checkUserActivityTime - checks manually calculated user activity with activity calculated by api.
func checkUserActivityTime(t *testing.T, client *http_client.NetHTTP, headers map[string]string, act []models.Activity) {
	log.Println("TEST: Requesting user activity from API")

	userID := act[0].UserID
	minTime, maxTime := getUserTiming(userID, act)

	var user = map[int64]bool{userID: true}
	activeTime, totalTime := manualActivityCalc(user, minTime, maxTime, act)

	path := fmt.Sprintf("http://localhost:9332/control/user/%d?TimeStart=%d&TimeEnd=%d",
		userID,
		minTime,
		maxTime,
	)

	bts, err := client.MakeRequest(http.MethodGet, path, headers, nil)

	if err != nil {
		t.Fatal(err)
	}

	data := new(models.DepartmentActivity)

	if err = json.Unmarshal(bts, &data); err != nil {
		t.Fatal(err)
	}

	checkTime(t, data.TotalTime, totalTime)
	checkTime(t, data.ActiveTime, activeTime)
}

// getDepUsers - separates user from given department.
func getDepUsers(depID int64, users []models.User) map[int64]bool {
	log.Println("TEST: Creating data time slice")

	usersMap := make(map[int64]bool)

	for _, u := range users {
		if u.DepartmentID == depID {
			usersMap[u.UserID] = true
		}
	}

	return usersMap
}

// getDepartTiming - get timing between first and last record, to check URL time query.
func getDepartTiming(depUsers map[int64]bool, activity []models.Activity) (int64, int64) {
	log.Println("TEST: Creating data time slice")

	var minTime, maxTime int64

	for _, a := range activity {
		if _, ok := depUsers[a.UserID]; !ok {
			continue
		}

		if a.Date > maxTime {
			maxTime = a.Date - 1
		}

		if a.Date > minTime {
			minTime = a.Date + 1
		}
	}

	return minTime, maxTime
}

// checkDepartActivityTime - checks manually calculated department activity with activity calculated by api.
func checkDepartActivityTime(t *testing.T,
	client *http_client.NetHTTP,
	headers map[string]string,
	act []models.Activity,
	users []models.User,
	depID int64,
) {
	log.Println("TEST: Requesting depart activity from API")

	depUsers := getDepUsers(depID, users)
	minTime, maxTime := getDepartTiming(depUsers, act)
	activeTime, totalTime := manualActivityCalc(depUsers, minTime, maxTime, act)

	path := fmt.Sprintf("http://localhost:9332/control/user/%d?TimeStart=%d&TimeEnd=%d",
		depID,
		minTime,
		maxTime,
	)

	bts, err := client.MakeRequest(http.MethodGet, path, headers, nil)

	if err != nil {
		t.Fatal(err)
	}

	data := new(models.UserActivity)

	if err = json.Unmarshal(bts, &data); err != nil {
		t.Fatal(err)
	}

	checkTime(t, data.TotalTime, totalTime)
	checkTime(t, data.ActiveTime, activeTime)
}

// checkTime - compares to times.
// First - time received from api.
func checkTime(t *testing.T, first *int64, second int64) {
	if first != nil {
		if *first != second {
			t.Fatalf("invalid time, expected: %d, got %d", second, *first)
		}

		return
	}

	if second != 0 {
		t.Fatalf("invalid time, expected: %d, got %d", second, 0)
	}
}

// deleteByIds - deletes given ids from given handler.
func deleteByIds(t *testing.T, client *http_client.NetHTTP, headers map[string]string, path string, ids []int64) {
	log.Println("TEST: Deleting objects by ID in path: " + path)

	for _, id := range ids {
		path := fmt.Sprintf("%s/%d", path, id)
		_, err := client.MakeRequest(http.MethodDelete, path, headers, nil)

		if err != nil {
			t.Fatal(err)
		}
	}
}

// deleteObjects - deletes all created objects.
func deleteObjects(t *testing.T,
	client *http_client.NetHTTP,
	headers map[string]string,
	deps []models.Department,
	users []models.User,
	activities []models.Activity,
) {
	log.Println("TEST: Deleting created objects...")

	depsID := make([]int64, len(deps))

	for id, d := range deps {
		depsID[id] = d.DepartmentID
	}

	deleteByIds(t, client, headers, "http://localhost:9332/departments", depsID)

	usersID := make([]int64, len(users))

	for id, u := range users {
		usersID[id] = u.UserID
	}

	deleteByIds(t, client, headers, "http://localhost:9332/users", usersID)

	actIds := make([]int64, len(activities))

	for id, act := range activities {
		actIds[id] = act.RecordID
	}

	deleteByIds(t, client, headers, "http://localhost:9332/activities", actIds)
}

// functionalCheck - base functional check for api.
// No need to return errors, could just t.Fatal(), to get full trace.
func functionalCheck(t *testing.T, wg *sync.WaitGroup) {
	log.Println("TEST: TEST: Starting check...")
	defer wg.Done()

	client := http_client.NewHTTPClient()

	userBts := getTestAdmin(t)
	headers := registerAndLogin(t, client, userBts)

	depIds := addDepartments(t, client, headers)
	users := addUsers(t, client, headers, depIds)
	activities := addActivities(t, client, headers, users)

	checkUserActivityTime(t, client, headers, activities)
	checkDepartActivityTime(t, client, headers, activities, users, depIds[0].DepartmentID)

	deleteObjects(t, client, headers, depIds, users, activities)
	unregister(t, client, headers, userBts)

	_, err := login(t, client, userBts)
	if err == nil {
		t.Fatal("User wasn't unregistered")
	}
}

// checkDeparts - checks if returned departs are equal to loaded departs.
func checkDeparts(t *testing.T, departs []models.Department, client *http_client.NetHTTP, headers map[string]string) {
	log.Println("Checking if returned departs are equal to loaded departs.")
	bts, err := client.MakeRequest(http.MethodGet, "http://localhost:9332/departments", headers, nil)

	if err != nil {
		t.Fatal(err)
	}

	data := make([]models.Department, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].DepartmentID < data[j].DepartmentID
	})

	sort.Slice(departs, func(i, j int) bool {
		return departs[i].DepartmentID < departs[j].DepartmentID
	})

	if !reflect.DeepEqual(departs, data) {
		t.Fatalf("Departments doesn't match, lenR: %d, lenL: %d", len(departs), len(data))
	}
}

// checkDeparts - checks if returned users are equal to loaded users.
func checkUsers(t *testing.T, users []models.User, client *http_client.NetHTTP, headers map[string]string) {
	log.Println("Checking if returned users are equal to loaded users.")
	bts, err := client.MakeRequest(http.MethodGet, "http://localhost:9332/users", headers, nil)

	if err != nil {
		t.Fatal(err)
	}

	data := make([]models.User, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].UserID < data[j].UserID
	})

	sort.Slice(users, func(i, j int) bool {
		return users[i].UserID < users[j].UserID
	})

	if !reflect.DeepEqual(users, data) {
		t.Fatalf("Users doesn't match, lenR: %d, lenL: %d", len(users), len(data))
	}
}

// checkDeparts - checks if returned activities are equal to loaded activities.
func checkActivities(t *testing.T, act []models.Activity, client *http_client.NetHTTP, headers map[string]string) {
	log.Println("Checking if returned activities are equal to loaded activities.")
	bts, err := client.MakeRequest(http.MethodGet, "http://localhost:9332/activities", headers, nil)

	if err != nil {
		t.Fatal(err)
	}

	data := make([]models.Activity, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].RecordID < data[j].RecordID
	})

	sort.Slice(act, func(i, j int) bool {
		return act[i].RecordID < act[j].RecordID
	})

	if !reflect.DeepEqual(act, data) {
		t.Fatalf("Activities doesn't match, lenR: %d, lenL: %d", len(act), len(data))
	}
}

// TODO: Requires clean database!!!!!!!!
// TODO: Mock database and implement unit tests
func getHandlersCheck(t *testing.T) {
	client := http_client.NewHTTPClient()

	userBts := getTestAdmin(t)
	headers := registerAndLogin(t, client, userBts)

	deps := addDepartments(t, client, headers)
	users := addUsers(t, client, headers, deps)
	activities := addActivities(t, client, headers, users)

	checkDeparts(t, deps, client, headers)
	checkUsers(t, users, client, headers)
	checkActivities(t, activities, client, headers)

	deleteObjects(t, client, headers, deps, users, activities)
	unregister(t, client, headers, userBts)
}

// Base functional test.
// No need to check delete, because getHandlersCheck will fail if functionalCheck doesn't delete it's data
func Test_AAPI(t *testing.T) {
	// Wait group to sync diff clients.
	var wg sync.WaitGroup
	srv := control.NewAAService(&config)
	go srv.Run()

	// Requires clean database!!!!!!!!
	for i := 0; i < 3; i++ {
		testName := fmt.Sprintf("Test: %d", i)
		wg.Add(1)

		// Run clients
		go t.Run(testName, func(t *testing.T) {
			functionalCheck(t, &wg)
		})
	}

	wg.Wait() // Wait for clients.

	// Requires clean database!!!!!!!!
	t.Run("GET handlers check", func(t *testing.T) {
		getHandlersCheck(t)
	})

	srv.Stop()
}
