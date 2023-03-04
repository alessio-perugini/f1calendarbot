package f1calendar

import (
	"fmt"
	"strings"
	"time"
)

type RaceWeek struct {
	Location string
	Round    int
	Sessions []SessionToBeDone
}

type SessionToBeDone struct {
	Name string
	Time time.Time
}

func (r RaceWeek) String() string {
	response := fmt.Sprintf("**%s** \n\n```\n", strings.ReplaceAll(r.Location, "-", "\\-"))

	for _, v := range r.Sessions {
		response += fmt.Sprintf("%-7s| %s\n", v.Name, v.Time.String())
	}
	response += "```"

	return response
}
