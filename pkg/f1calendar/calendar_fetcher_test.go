package f1calendar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shoenig/test/must"
)

func TestCalendarFetcher_GetRaceWeek(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 14, 12, 0, 0, 0, time.UTC)
	rome, err := time.LoadLocation("Europe/Rome")
	must.NoError(t, err)

	const past = `{"location":"Old","round":1,"sessions":{"gp":"2026-03-01T14:00:00Z"}}`
	const next = `{"location":"Las Vegas","round":22,"sessions":{"gp":"2026-11-22T04:00:00Z","fp1":"2026-11-20T00:30:00Z","qualifying":"2026-11-21T04:00:00Z","bad":"not-a-time"}}`
	const nextSeason = `{"location":"Melbourne","round":1,"sessions":{"fp1":"2027-03-05T01:30:00Z"}}`

	tests := []struct {
		name      string
		responses map[string]response
		want      *RaceWeek
		wantErr   bool
	}{
		{
			name:      "next race of current season, sorted, invalid session skipped",
			responses: map[string]response{"/2026.json": {200, `{"races":[` + past + `,` + next + `]}`}},
			want: &RaceWeek{Location: "Las Vegas", Round: 22, Sessions: []SessionToBeDone{
				{Name: "fp1", Time: time.Date(2026, 11, 20, 0, 30, 0, 0, time.UTC).In(rome)},
				{Name: "qualifying", Time: time.Date(2026, 11, 21, 4, 0, 0, 0, time.UTC).In(rome)},
				{Name: "gp", Time: time.Date(2026, 11, 22, 4, 0, 0, 0, time.UTC).In(rome)},
			}},
		},
		{
			name: "season over falls back to next season",
			responses: map[string]response{
				"/2026.json": {200, `{"races":[` + past + `]}`},
				"/2027.json": {200, `{"races":[` + nextSeason + `]}`},
			},
			want: &RaceWeek{Location: "Melbourne", Round: 1, Sessions: []SessionToBeDone{
				{Name: "fp1", Time: time.Date(2027, 3, 5, 1, 30, 0, 0, time.UTC).In(rome)},
			}},
		},
		{
			name:      "season over and next season not published",
			responses: map[string]response{"/2026.json": {200, `{"races":[` + past + `]}`}},
		},
		{
			name:      "server error",
			responses: map[string]response{"/2026.json": {http.StatusServiceUnavailable, "oops"}},
			wantErr:   true,
		},
		{
			name:      "invalid json",
			responses: map[string]response{"/2026.json": {200, "<html>"}},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp, ok := tt.responses[r.URL.Path]
				if !ok {
					resp = response{http.StatusNotFound, "404: Not Found"}
				}
				w.WriteHeader(resp.status)
				_, _ = w.Write([]byte(resp.body))
			}))
			t.Cleanup(srv.Close)

			f := NewCalendarFetcher(srv.URL, rome)
			f.now = func() time.Time { return now }

			got, err := f.GetRaceWeek(context.Background())
			if tt.wantErr {
				must.Error(t, err)
				return
			}
			must.NoError(t, err)
			must.Eq(t, tt.want, got)
		})
	}
}

type response struct {
	status int
	body   string
}
