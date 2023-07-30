package f1calendar

import (
	"fmt"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	tw := table.NewWriter()
	tw.SetTitle(r.Location + " (TZ: UTC+2)")
	tw.AppendHeader(table.Row{"Session", "Time"})
	tw.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
		{Number: 2, AlignFooter: text.AlignCenter, AlignHeader: text.AlignCenter},
	})

	for _, v := range r.Sessions {
		switch v.Name {
		case "qualifying":
			v.Name = "ql"
		case "sprintQualifying":
			v.Name = "sql"
		}
		tw.AppendRow(table.Row{v.Name, v.Time.Format("2006-01-02 15:04:05")})
	}

	return fmt.Sprintf("<pre>%s</pre>", tw.Render())
}
