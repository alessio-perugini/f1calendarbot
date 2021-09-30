package domain

import "time"

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
