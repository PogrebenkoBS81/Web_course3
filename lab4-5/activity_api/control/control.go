package control

import (
	"activity_api/api"
	"activity_api/common/cancellation"
	"activity_api/data_manager/cache"
	"activity_api/data_manager/db"
	"activity_api/data_manager/db/core"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// TODO: implement HTTPS. I already have functions for JWT to generate valid keys, so in future I need to get .crt
// AAService - config for service
type AAService struct {
	addr string // addr of service

	api   *api.AApi           // service api
	cache cache.ICacheManager // used for storing tokens in auth
	db    core.ISQLDatabase   // SQL db for user data

	logger logrus.FieldLogger
	cancel *cancellation.Token // service cancellation token for cancel management
	wg     sync.WaitGroup      // used to wait for all processes to finish
}

// NewAAService - returns new AAService with given parameters
func NewAAService(config *AAServiceConfig) *AAService {
	logger := &logrus.Logger{
		Formatter: new(logrus.TextFormatter),
		Out:       os.Stdout,
		Level:     logrus.Level(config.LogLevel),
	}

	aaService := &AAService{
		addr: config.Addr,
		cancel: cancellation.NewCustomToken(
			context.Background(),
			pingersNum, // pingersNum - number of pingers that will ping IManageable services
		),
		db:     db.NewAADatabase(config.DbType, config.ConnString, logger),
		logger: logger.WithField("module", "AAService"),
	}

	aaService.cache = cache.NewCacheManager(
		config.CacheType,
		config.Cache,
		aaService.cancel.Context(),
		logger,
	)

	aaService.api = api.NewAApi(config.Addr, aaService.db, aaService.cache, aaService.cancel.Context(), logger)

	return aaService
}

// stop - stops AAService modules.
func (a *AAService) stop() (err error) {
	a.logger.WithField("func", "stop").Info("Stopping AAService modules... ")

	if err = a.api.Close(); err != nil {
		err = fmt.Errorf("AAService api Close(): %w", err)
	}

	if err = a.db.Close(); err != nil {
		err = fmt.Errorf("AAService db Close(): %w", err)
	}

	if err = a.cache.Close(); err != nil {
		err = fmt.Errorf("AAService cache Close(): %w", err)
	}

	return
}

// Stop - stops AAService and waits for all processes to finish.
func (a *AAService) Stop() {
	a.logger.WithField("func", "Stop").Info("Starting AAService... ")

	// Could be called multiple times.
	// There is no need in cancellation token here,
	// because this Start and Stop would be called only from 1 routine only in open->close order,
	// but in possible future it could help to organize project further.
	a.cancel.Cancel().Await()
	a.wg.Wait()
}

// Run - prepares DB and Redis to work, starts pingers for them,
// opens api for requests and then waits for signal to stop.
func (a *AAService) Run() {
	a.wg.Add(1)
	defer a.wg.Done()

	entry := a.logger.WithField("func", "Run")
	entry.Info("Starting...")
	// Open db and create required tables for database if they don't exist.
	if err := a.initDatabase(); err != nil {
		entry.Fatalf("database init error: %v", err)
	}
	// Open redis connection.
	if err := a.cache.Open(); err != nil {
		entry.Fatalf("cache open error: %v", err)
	}

	go a.pinger(a.db)    // Start pinger for db
	go a.pinger(a.cache) // Start pinger for redis
	go a.api.Start()
	// Chanel for signal to wait interrupt, so ctrl + c could stop service.
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		entry.Info("Incoming signal:", <-signals)
		a.cancel.Cancel() // Cancel all processes
	}()

	entry.Info("Awaiting stop signal...")
	a.cancel.Await() // Wait for all services to stop.
	entry.Info("Exiting")

	if err := a.stop(); err != nil {
		a.logger.Errorf("stop AAService error: %v", err)
	}
}

// initDatabase - opens db connection, and creates required tables for database if they don't exist.
func (a *AAService) initDatabase() error {
	a.logger.WithField("func", "initDatabase").Info("Preparing DB to work...")

	if err := a.db.Open(); err != nil {
		return fmt.Errorf("AAService db.Open(): %w", err)
	}
	// Create missing tables
	if err := a.db.CreateDB(); err != nil {
		return fmt.Errorf("AAService db.Create(): %w", err)
	}

	return nil
}

// pinger - pings given service and tries to restart it if it crashes.
func (a *AAService) pinger(service iManageable) {
	entry := a.logger.WithField("func", "pinger")
	entry.Info("Starting pinger for service:", service.Describe())

	defer a.cancel.Done()
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-a.cancel.Cancelled(): // cancel pinger when program is stopping.
			entry.Info("Stopping for:", service.Describe())

			return
		case <-ticker.C: // ping db every tick.
			a.ping(service)
		}
	}
}

// ping - checks service for connection, tries to reconnect if it crashes.
func (a *AAService) ping(service iManageable) {
	entry := a.logger.WithField("func", "ping")

	entry.Info("Doing ping for service:", service.Describe())
	err := service.OK()

	if err == nil {
		return
	}

	entry.Errorf("Ping %s, OK() error: %v", service.Describe(), err)
	entry.Info("Trying to reconnect: ", service.Describe())

	if err = service.Restart(); err != nil {
		entry.Errorf("Ping %s, Restart() error: %v", service.Describe(), err)

		return
	}

	entry.Infof("Reconnect to %s successful!", service.Describe())
}
