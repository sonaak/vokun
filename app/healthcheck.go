package app

import (
	"time"
	"sync"
	"github.com/golang/glog"
)

type status string

const (
	OK = "ok"
	ERROR = "error"
)

type DbStatus struct {
	Connections int `json:"connections"`
	RoundTrip float64 `json:"roundtrip"`
	ErrorRate float64 `json:"error_rate"`
}

type HealthCheckStats struct {
	Status status `json:"status"`
	Db DbStatus `json:"db"`
}


type HealthChecker struct {}


func NewHealthChecker() *HealthChecker {
	return &HealthChecker{}
}


func (hc *HealthChecker) checkDb(db SqlClient) DbStatus {
	// gather the stats about open connections
	stats := db.Stats()

	// ping it 10 times and aggregate the results
	pingChan := make(chan struct { Err error; Duration time.Duration }, 10)
	wg := sync.WaitGroup {}
	for i := 0; i < 10; i ++ {
		wg.Add(1)

		go func() {
			startTime := time.Now()
			pingErr := db.Ping()
			if pingErr != nil {
				glog.Errorf("Error pinging db: %v", pingErr)
			}
			pingChan <- struct {
				Err error
				Duration time.Duration
			}{
				Err: pingErr,
				Duration: time.Since(startTime),
			}
			wg.Done()
		}()
	}
	wg.Wait()
	close(pingChan)

	connSpeed := 0.00
	errorRate := 0.00
	for s := range pingChan {
		if s.Err != nil {
			errorRate += 0.1
		}

		connSpeed += s.Duration.Seconds() * 100
	}

	return DbStatus {
		Connections: stats.OpenConnections,
		ErrorRate: errorRate,
		RoundTrip: connSpeed,
	}
}


func (hc *HealthChecker) CheckHealth(db SqlClient) HealthCheckStats {
	// check for a total of 10 times
	dbStats := hc.checkDb(db)

	return HealthCheckStats {
		Status: OK,
		Db: dbStats,
	}
}
