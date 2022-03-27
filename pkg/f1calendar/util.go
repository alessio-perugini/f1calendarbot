package f1calendar

import (
	"time"

	"github.com/rs/zerolog/log"
)

func mustParseTime(t string) time.Time {
	result, err := time.Parse(time.RFC3339, t)
	if err != nil {
		log.Fatal().Msgf("Unable to parse fp1 datetime %v", err)
	}

	return result.In(time.Local)
}
