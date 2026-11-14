package f1calendar

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var sessionNames = map[string]string{
	"fp1":              "Practice 1",
	"fp2":              "Practice 2",
	"fp3":              "Practice 3",
	"qualifying":       "Qualifying",
	"sprintQualifying": "Sprint Qualifying",
	"sprint":           "Sprint",
	"gp":               "Race",
}

type RaceWeek struct {
	Location string
	Round    int
	Sessions []SessionToBeDone
}

type SessionToBeDone struct {
	Name string
	Time time.Time
}

// DisplayName returns a human readable session name.
func (s SessionToBeDone) DisplayName() string {
	if n, ok := sessionNames[s.Name]; ok {
		return n
	}
	return s.Name
}

// upcoming returns a copy of the race week with only the sessions starting after now,
// or nil when none is left.
func (r *RaceWeek) upcoming(now time.Time) *RaceWeek {
	if r == nil {
		return nil
	}
	sessions := make([]SessionToBeDone, 0, len(r.Sessions))
	for _, s := range r.Sessions {
		if now.Before(s.Time) {
			sessions = append(sessions, s)
		}
	}
	if len(sessions) == 0 {
		return nil
	}
	return &RaceWeek{Location: r.Location, Round: r.Round, Sessions: sessions}
}

func (r RaceWeek) String() string {
	tw := table.NewWriter()
	tw.SetTitle(r.Location + " (TZ: Europe/Rome)")
	tw.AppendHeader(table.Row{"Session", "Time"})
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 2, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
	})

	for _, v := range r.Sessions {
		tw.AppendRow(table.Row{v.DisplayName(), v.Time.Format("Mon 02 Jan 15:04")})
	}

	return fmt.Sprintf("<pre>%s</pre>", tw.Render())
}
