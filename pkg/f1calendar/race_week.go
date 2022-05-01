package f1calendar

import (
	"fmt"
	"time"
)

type RaceWeekRepository interface {
	GetRaceWeek() *RaceWeek
}

type RaceWeek struct {
	Location string
	Round    int
	Sessions []*Session
}

type Session struct {
	Name string
	Time time.Time
}

func (r *RaceWeek) String() string {
	response := fmt.Sprintf("%s \n\n```\n", r.Location)

	for _, v := range r.Sessions {
		response += fmt.Sprintf("%-7s| %s\n", v.Name, v.Time.String())
	}
	response += fmt.Sprintf("```")

	return response
}
