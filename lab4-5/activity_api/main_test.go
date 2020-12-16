// Basic smoke test
// This test is quite messy, so it would be better to rewrite them in future.
// TODO: Clean up this test

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
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

// Predefined config for AAService
var config = control.AAServiceConfig{
	CacheType: 0, // CacheMock
	DbType:    0,
	// // Will be generated randomly to ensure clean DB
	// ConnString: "functional_test_db.db",
	Addr:     "localhost:9332",
	LogLevel: 4, // Info level
	// Cache Mock doesn't need config
	Cache: &cache.ICacheConfig{},
}

type testRunner func(data *loadData)

type loadData struct {
	deps    []*models.Department
	users   []*models.User
	act     []*models.Activity
	headers map[string]string
}

// smokeTest - base smoke test realisation
type smokeTest struct {
	client *http_client.NetHTTP
	t      *testing.T
}

// newSmokeTest - create new smoke tester
func newSmokeTest(client *http_client.NetHTTP, t *testing.T) *smokeTest {
	return &smokeTest{
		client: client,
		t:      t,
	}
}

// getTestAdmin - returns randomly created admin.
func (s *smokeTest) getTestAdmin() []byte {
	log.Println("TEST: TEST: getting test admin")

	u := models.Admin{
		Username: uuid.New().String(),
		// I didn't truly create hash here, because it's only a test,
		// but if sometime client would be implemented - has should be calculated on the client side.
		Hash: uuid.New().String(),
	}

	bts, err := json.Marshal(u)

	if err != nil {
		s.t.Fatal(err)
	}

	return bts
}

// login - attempts to login with given admin.
func (s *smokeTest) login(user []byte) (map[string]string, error) {
	log.Println("TEST: log in API")

	bts, err := s.client.MakeRequest(http.MethodPost, "http://localhost:9332/login", nil, user)

	if err != nil {
		return nil, err
	}

	m := make(map[string]string)

	if err = json.Unmarshal(bts, &m); err != nil {
		s.t.Fatal(err)
	}

	return m, nil
}

// postObject - posts given data on given handler.
func (s *smokeTest) postObject(headers map[string]string, path string, data interface{}) int64 {
	bts, err := json.Marshal(data)
	if err != nil {
		s.t.Fatal(err)
	}

	bts, err = s.client.MakeRequest(http.MethodPost, path, headers, bts)
	if err != nil {
		s.t.Fatal(err)
	}

	id := new(models.ObjectID)

	if err = json.Unmarshal(bts, &id); err != nil {
		s.t.Fatal(err)
	}

	return id.ID
}

// registerAndLogin - register given admin and logins it, returns auth token
func (s *smokeTest) registerAndLogin(userBts []byte) map[string]string {
	log.Println("TEST: Registering and login")

	_, err := s.client.MakeRequest(http.MethodPost, "http://localhost:9332/register", nil, userBts)

	if err != nil {
		s.t.Fatal(err)
	}

	token, err := s.login(userBts)
	if err != nil {
		s.t.Fatal(err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + token["access_token"],
	}

	return headers
}

// unregister - unregister given admin.
func (s *smokeTest) unregister(headers map[string]string, userBts []byte) {
	log.Println("TEST: Unregistering")
	_, err := s.client.MakeRequest(http.MethodDelete, "http://localhost:9332/unregister", headers, userBts)

	if err != nil {
		s.t.Fatal(err)
	}
}

// addDepartments - adds randomly generated departments.
func (s *smokeTest) addDepartments(headers map[string]string) []*models.Department {
	log.Println("TEST: Adding random departments")
	deps := make([]*models.Department, 0)

	for i := 0; i < rand.Intn(100); i++ {
		dep := &models.Department{
			DepartmentName: uuid.New().String(),
		}

		id := s.postObject(headers, "http://localhost:9332/departments", dep)
		dep.DepartmentID = id
		deps = append(deps, dep)
	}

	return deps
}

// addUsers - adds randomly generated users to given departments.
func (s *smokeTest) addUsers(headers map[string]string, deps []*models.Department) []*models.User {
	log.Println("TEST: Adding random users")
	users := make([]*models.User, 0)

	for _, dep := range deps {
		for i := 0; i < rand.Intn(10); i++ {
			user := &models.User{
				UserName:     uuid.New().String(),
				DepartmentID: dep.DepartmentID,
			}

			id := s.postObject(headers, "http://localhost:9332/users", user)

			user.UserID = id
			users = append(users, user)
		}
	}

	return users
}

// addActivities - adds randomly generated activities to given users.
func (s *smokeTest) addActivities(headers map[string]string, users []*models.User) []*models.Activity {
	log.Println("TEST: Adding random activities")

	activities := make([]*models.Activity, 0)
	var min int64 = 500
	var max int64 = 1000

	for _, u := range users {
		for i := 0; i < rand.Intn(100); i++ {
			user := &models.Activity{
				UserID:     u.UserID,
				TotalTime:  rand.Int63n(max-min) + min,
				ActiveTime: rand.Int63n(min),
				Date:       time.Now().Unix() + rand.Int63n(max),
			}

			id := s.postObject(headers, "http://localhost:9332/activities", user)

			user.RecordID = id
			activities = append(activities, user)
		}
	}

	return activities
}

// manualCalc - manually calculates activity time for given users to check api.
func (s *smokeTest) manualTimeCalc(act []*models.Activity, users map[int64]bool, minTime, maxTime int64) (int64, int64) {
	log.Println("TEST: Manually calculating activity")

	var manualActive int64
	var manualTotal int64

	for _, a := range act {
		if _, ok := users[a.UserID]; !ok {
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
func (s *smokeTest) getUserTiming(act []*models.Activity, userID int64) (int64, int64) {
	log.Println("TEST: Creating data time slice")

	sliceMinTime := act[0].Date + 1 // Don't include first user record to test URL time query
	sliceMaxTime := act[0].Date

	for _, a := range act {
		if a.UserID == userID && sliceMaxTime < a.Date {
			sliceMaxTime = a.Date - 1 // Don't include last user record to test URL time query
		}
	}

	return sliceMinTime, sliceMaxTime
}

// checkUserActivityTime - checks manually calculated user activity with activity calculated by api.
func (s *smokeTest) checkUserActivityTime(act []*models.Activity, headers map[string]string) {
	log.Println("TEST: Requesting user activity from API")

	userID := act[0].UserID
	minTime, maxTime := s.getUserTiming(act, userID)

	var user = map[int64]bool{userID: true}
	activeTime, totalTime := s.manualTimeCalc(act, user, minTime, maxTime)

	path := fmt.Sprintf("http://localhost:9332/control/user/%d?TimeStart=%d&TimeEnd=%d",
		userID,
		minTime,
		maxTime,
	)

	bts, err := s.client.MakeRequest(http.MethodGet, path, headers, nil)

	if err != nil {
		s.t.Fatal(err)
	}

	data := new(models.DepartmentActivity)

	if err = json.Unmarshal(bts, &data); err != nil {
		s.t.Fatal(err)
	}

	s.checkTime(data.TotalTime, totalTime)
	s.checkTime(data.ActiveTime, activeTime)
}

// getDepUsers - separates user from given department.
func (s *smokeTest) getDepUsers(users []*models.User, depID int64) map[int64]bool {
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
func (s *smokeTest) getDepartTiming(act []*models.Activity, depUsers map[int64]bool) (int64, int64) {
	log.Println("TEST: Creating data time slice")

	var minTime, maxTime int64

	for _, a := range act {
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

// manualDepartCalc - checks manually calculated department activity with activity calculated by api.
func (s *smokeTest) checkDepartActivityTime(
	act []*models.Activity,
	users []*models.User,
	depID int64,
	headers map[string]string,
) {
	log.Println("TEST: Requesting depart activity from API")

	depUsers := s.getDepUsers(users, depID)
	minTime, maxTime := s.getDepartTiming(act, depUsers)
	activeTime, totalTime := s.manualTimeCalc(act, depUsers, minTime, maxTime)

	path := fmt.Sprintf("http://localhost:9332/control/user/%d?TimeStart=%d&TimeEnd=%d",
		depID,
		minTime,
		maxTime,
	)

	bts, err := s.client.MakeRequest(http.MethodGet, path, headers, nil)

	if err != nil {
		s.t.Fatal(err)
	}

	data := new(models.UserActivity)

	if err = json.Unmarshal(bts, &data); err != nil {
		s.t.Fatal(err)
	}

	s.checkTime(data.TotalTime, totalTime)
	s.checkTime(data.ActiveTime, activeTime)
}

// checkTime - compares to times.
// First - time received from api.
func (s *smokeTest) checkTime(first *int64, second int64) {
	if first != nil {
		if *first != second {
			s.t.Fatalf("invalid time, expected: %d, got %d", second, *first)
		}

		return
	}

	if second != 0 {
		s.t.Fatalf("invalid time, expected: %d, got %d", second, 0)
	}
}

// deleteByIds - deletes given ids from given handler.
func (s *smokeTest) deleteByIds(headers map[string]string, path string, ids []int64) {
	log.Println("TEST: Deleting objects by ID in path: " + path)

	for _, id := range ids {
		path := fmt.Sprintf("%s/%d", path, id)
		_, err := s.client.MakeRequest(http.MethodDelete, path, headers, nil)

		if err != nil {
			s.t.Fatal(err)
		}
	}
}

// deleteObjects - deletes all created objects.
func (s *smokeTest) deleteObjects(ld *loadData) {
	log.Println("TEST: Deleting created objects...")

	depsID := make([]int64, len(ld.deps))

	for id, d := range ld.deps {
		depsID[id] = d.DepartmentID
	}

	s.deleteByIds(ld.headers, "http://localhost:9332/departments", depsID)

	usersID := make([]int64, len(ld.users))

	for id, u := range ld.users {
		usersID[id] = u.UserID
	}

	s.deleteByIds(ld.headers, "http://localhost:9332/users", usersID)

	actIds := make([]int64, len(ld.act))

	for id, act := range ld.act {
		actIds[id] = act.RecordID
	}

	s.deleteByIds(ld.headers, "http://localhost:9332/activities", actIds)
}

// checkDeparts - checks if returned departs are equal to loaded departs.
func (s *smokeTest) checkDeparts(deps []*models.Department, headers map[string]string) {
	log.Println("Checking if returned departs are equal to loaded departs.")
	bts, err := s.client.MakeRequest(http.MethodGet, "http://localhost:9332/departments", headers, nil)

	if err != nil {
		s.t.Fatal(err)
	}

	data := make([]*models.Department, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		s.t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].DepartmentID < data[j].DepartmentID
	})

	sort.Slice(deps, func(i, j int) bool {
		return deps[i].DepartmentID < deps[j].DepartmentID
	})

	if !reflect.DeepEqual(deps, data) {
		s.t.Fatalf("Departments doesn't match, lenRecieved: %d, lenLoaded: %d", len(data), len(deps))
	}
}

// checkDeparts - checks if returned users are equal to loaded users.
func (s *smokeTest) checkUsers(users []*models.User, headers map[string]string) {
	log.Println("Checking if returned users are equal to loaded users.")
	bts, err := s.client.MakeRequest(http.MethodGet, "http://localhost:9332/users", headers, nil)

	if err != nil {
		s.t.Fatal(err)
	}

	data := make([]*models.User, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		s.t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].UserID < data[j].UserID
	})

	sort.Slice(users, func(i, j int) bool {
		return users[i].UserID < users[j].UserID
	})

	if !reflect.DeepEqual(users, data) {
		s.t.Fatalf("Users doesn't match, lenRecieved: %d, lenLoaded: %d", len(data), len(users))
	}
}

// checkDeparts - checks if returned activities are equal to loaded activities.
func (s *smokeTest) checkActivities(act []*models.Activity, headers map[string]string) {
	log.Println("Checking if returned activities are equal to loaded activities.")
	bts, err := s.client.MakeRequest(http.MethodGet, "http://localhost:9332/activities", headers, nil)

	if err != nil {
		s.t.Fatal(err)
	}

	data := make([]*models.Activity, 0)

	if err = json.Unmarshal(bts, &data); err != nil {
		s.t.Fatal(err)
	}

	sort.Slice(data, func(i, j int) bool {
		return data[i].RecordID < data[j].RecordID
	})

	sort.Slice(act, func(i, j int) bool {
		return act[i].RecordID < act[j].RecordID
	})

	if !reflect.DeepEqual(act, data) {
		s.t.Fatalf("Activities doesn't match, lenRecieved: %d, lenLoaded: %d", len(data), len(act))
	}
}

// checkDelete - checks if all data were successfully deleted
func (s *smokeTest) checkDelete(headers map[string]string) {
	log.Println("Checking for deletion...")

	s.checkDeparts(make([]*models.Department, 0), headers)
	s.checkUsers(make([]*models.User, 0), headers)
	s.checkActivities(make([]*models.Activity, 0), headers)
}

// TestRunner - generates random data, passes it to test scenario, and executes it
// Deletes all data afterwards
func (s *smokeTest) TestRunner(testCase testRunner) {
	log.Println("TEST: Starting test runner...")

	userBts := s.getTestAdmin()
	headers := s.registerAndLogin(userBts)

	deps := s.addDepartments(headers)
	users := s.addUsers(headers, deps)
	activities := s.addActivities(headers, users)

	ld := &loadData{
		deps:    deps,
		users:   users,
		act:     activities,
		headers: headers,
	}

	testCase(ld)

	s.deleteObjects(ld)
	s.unregister(headers, userBts)

	_, err := s.login(userBts)
	if err == nil {
		s.t.Fatal("User wasn't unregistered")
	}
}

// TestTimeCalc - checks time calculations.
// No need to return errors, could just t.Fatal(), to get full trace.
func (s *smokeTest) TestTimeCalc(ld *loadData) {
	log.Println("TEST: Starting time test...")

	s.checkUserActivityTime(ld.act, ld.headers)
	s.checkDepartActivityTime(ld.act, ld.users, ld.deps[0].DepartmentID, ld.headers)
}

// TestGet - tests all GET handlers.
// No need to return errors, could just t.Fatal(), to get full trace.
func (s *smokeTest) TestGet(ld *loadData) {
	log.Println("TEST: Starting GET/DELETE check...")

	s.checkDeparts(ld.deps, ld.headers)
	s.checkUsers(ld.users, ld.headers)
	s.checkActivities(ld.act, ld.headers)
}

// RunMultiple - allows to wait for multiple routines to exit
func (s *smokeTest) RunMultiple(wg *sync.WaitGroup, testCase testRunner) {
	defer wg.Done()
	s.TestRunner(testCase)
}

func RunSmokeTest(t *testing.T) {
	// Wait group to sync diff clients.
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		testName := fmt.Sprintf("Smoke_test_%d", i)
		client := http_client.NewHTTPClient()

		wg.Add(1)

		// Run clients
		go t.Run(testName, func(t *testing.T) {
			test := newSmokeTest(client, t)
			test.RunMultiple(&wg, test.TestTimeCalc)
		})
	}
	// Wait for clients.
	wg.Wait()

	// No need for DELETE test, if not all objects were deleted - this test fill fall
	t.Run("DET/DELETE_test", func(t *testing.T) {
		test := newSmokeTest(http_client.NewHTTPClient(), t)
		test.TestRunner(test.TestGet)
	})
}

// Base smoke test.
func Test_AAPI(t *testing.T) {
	// Since conn string for SQLite is it's path - generate tmp db in curr dir, and delete it after test
	config.ConnString = uuid.New().String() + ".db"
	defer func() {
		if err := os.Remove(config.ConnString); err != nil {
			t.Error(err)
		}
	}()
	// Run service for test
	srv := control.NewAAService(&config)
	go srv.Run()

	RunSmokeTest(t)

	srv.Stop()
}
