package server

import (
	"log"
	"time"
)

// TimeManager - time manager to process time.
// Its called only in doOnce, and client manager with mutex
// So i dont think that sync should be here.
type TimeManager struct {
	interval int
	started  time.Time
	ticker   *time.Ticker
}

// newTimeManager - returns a new time manager.
func newTimeManager(interval int) *TimeManager {
	return &TimeManager{
		interval: interval,
		started:  time.Now(),
		ticker:   new(time.Ticker),
	}
}

// Starts ticker, and writes time of its start.
func (t *TimeManager) startTimer() {
	log.Println("Starting ticker with interval:", t.interval)
	t.ticker = time.NewTicker(time.Second * time.Duration(t.interval))
	t.started = time.Now()
}

// Returns time from timer start.
func (t *TimeManager) getTime() int64 {
	return t.started.Unix()
}
