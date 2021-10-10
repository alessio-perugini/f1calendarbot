package domain

import (
	"fmt"
	"time"
)

type RaceWeek struct {
	Location string
	Round    int
	Sessions []*Session
}

type Session struct {
	Name string
	Time time.Time
}

type F1RaceWeeRepository interface {
	GetCalendar() *RaceWeek
}

func (r *RaceWeek) String() string {
	response := fmt.Sprintf("%s \n\n", r.Location)

	for _, v := range r.Sessions {
		response += fmt.Sprintf("%s: %s\n", v.Name, v.Time.String())
	}

	return response
}
