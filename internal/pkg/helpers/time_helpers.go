package helpers

import (
	"time"

	"github.com/rs/zerolog/log"
)

// ParseDuration parses a duration string, returns default duration on error.
// Moved from internal/config package.
func ParseDuration(durationStr string, defaultDuration time.Duration) time.Duration {
	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		// Use the global logger here, assuming logger might not be configured when this is called.
		log.Warn().Err(err).Str("durationStr", durationStr).Dur("defaultDuration", defaultDuration).Msg("Failed to parse duration string, using default")
		return defaultDuration
	}
	return duration
}
