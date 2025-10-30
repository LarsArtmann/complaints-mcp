// Package cast provides type casting utilities
package cast

import (
	"time"
)

// ToDuration converts time.Duration to milliseconds for API responses
func ToDuration(d time.Time) float64 {
	return float64(d.Nanosecond()) / float64(time.Millisecond)
}