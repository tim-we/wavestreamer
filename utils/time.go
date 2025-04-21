package utils

import (
	"fmt"
	"time"
)

func PrettyDuration(d time.Duration, suffix string) string {
	seconds := int(d.Seconds())
	minutes := int(d.Minutes())
	hours := int(d.Hours())

	switch {
	case seconds < 60:
		return "just now"
	case minutes == 1:
		return fmt.Sprintf("1 minute%s", suffix)
	case minutes < 60:
		return fmt.Sprintf("%d minutes%s", minutes, suffix)
	case hours == 1:
		return fmt.Sprintf("1 hour%s", suffix)
	case hours < 24:
		return fmt.Sprintf("%d hours%s", hours, suffix)
	case hours < 48:
		return "yesterday"
	default:
		return fmt.Sprintf("%d days%s", hours/24, suffix)
	}
}
