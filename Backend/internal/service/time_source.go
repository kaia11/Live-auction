package service

import (
	"time"

	"auction-live/backend/internal/realtime"
)

func nowUnix(runtime *realtime.Runtime) int64 {
	if runtime != nil {
		if ts, err := runtime.NowUnix(); err == nil && ts > 0 {
			return ts
		}
	}
	return time.Now().Unix()
}

func nowTime(runtime *realtime.Runtime) time.Time {
	return time.Unix(nowUnix(runtime), 0).UTC()
}
