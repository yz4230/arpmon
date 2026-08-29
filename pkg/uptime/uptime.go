package uptime

import "time"

var startTime time.Time

func Initialize() {
	startTime = time.Now()
}

func Get() time.Duration {
	return time.Since(startTime)
}

func GetStartTime() time.Time {
	return startTime
}
