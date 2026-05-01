package ntputils

import (
	"time"
	"fmt"
	"strings"
)

func formatUnit(duration time.Duration) string {
	if duration < 100*time.Nanosecond {
		return fmt.Sprintf("%d ns", duration.Nanoseconds())
	}

	if duration < 100*time.Microsecond {
		return fmt.Sprintf("%d µs", duration.Microseconds())
	}

	return fmt.Sprintf("%.1f ms", float64(duration)/float64(time.Millisecond))
}

func formatPrecision(duration time.Duration, precision_unit string) string {
	if precision_unit == "ns" || precision_unit == "nanoseconds" {
		return fmt.Sprintf("%d", duration.Nanoseconds())
	}
	if precision_unit == "us" || precision_unit == "microseconds" {
		return fmt.Sprintf("%d", duration.Microseconds())
	}
	return fmt.Sprintf("%.3f", float64(duration)/float64(time.Millisecond))
}

func FormatDuration(duration time.Duration, precision_unit string) string {
	if duration < 0 {
		duration = -duration
	}
	if precision_unit == "" || precision_unit == "auto" {
		return formatUnit(duration)
	}
	return formatPrecision(duration, precision_unit)
}

func FormatTime(t time.Time, timestamp string, format string) string {
	if strings.ToLower(timestamp) == "true" {
		return fmt.Sprintf("%d", t.Unix())
	}
	if format == "" {
		format = "2006-01-02T15:04:05.000Z"
	}
	return t.Format(format)
}
