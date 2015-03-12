// Package hitcounter augments the message-server with a store to track hits.
package hitcounter

import (
	"fmt"
	"github.com/BruteForceFencer/core/logger"
	"github.com/BruteForceFencer/core/message-server"
	"os"
	"time"
)

// HitCounter is a server that tracks several directions.
type HitCounter struct {
	Clock  *Clock
	Count  *RunningCount
	Logger *logger.Logger
	*server.Server
}

// NewHitCounter returns an initialized *HitCounter.
func NewHitCounter(directions []Direction) *HitCounter {
	result := new(HitCounter)
	result.Clock = NewClock()
	result.Count = NewRunningCount(128, 24*time.Hour)
	result.Logger = logger.New(os.Stdout)
	result.Server = server.New()

	for i := range directions {
		// Add the route
		result.Routes[directions[i].Name] = makeRoute(result, &directions[i])

		// Schedule the cleanup
		go func(dir *Direction) {
			for {
				dir.Store.CleanUp(result.Clock.GetTime())
				time.Sleep(time.Duration(dir.CleanUpTime) * time.Second)
			}
		}(&directions[i])
	}

	return result
}

func makeRoute(hitCounter *HitCounter, dir *Direction) func(interface{}) bool {
	return func(val interface{}) bool {
		hitCounter.Count.Inc()
		valid := dir.Hit(hitCounter.Clock.GetTime(), val)
		if !valid {
			hitCounter.Logger.Log(dir.Name, fmt.Sprint(val))
		}

		return valid
	}
}
