package plugins

import (
	"time"
)

func TimeUnixMilli(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond/time.Nanosecond)
}

func DurationMilliseconds(d time.Duration) int64 {
	return d.Nanoseconds() / int64(time.Millisecond/time.Nanosecond)
}
