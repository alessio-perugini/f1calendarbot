package f1calendar

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

const peerName = "github-sportstimes"

type CalendarFetcher struct {
	client  *http.Client
	baseURL string
	loc     *time.Location
	now     func() time.Time
}

// NewCalendarFetcher fetches the season calendar from baseURL/<year>.json
// and reports session times in loc.
func NewCalendarFetcher(baseURL string, loc *time.Location) *CalendarFetcher {
	return &CalendarFetcher{
		client: &http.Client{
			Timeout: 15 * time.Second,
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
				otelhttp.WithSpanOptions(
					trace.WithAttributes(semconv.ServicePeerName(peerName)),
				),
			),
		},
		baseURL: baseURL,
		loc:     loc,
		now:     time.Now,
	}
}

// GetRaceWeek returns the next race week with upcoming sessions, looking at the
// current season first and at the next one once the current season is over.
// It returns nil, nil when no race is scheduled.
func (c *CalendarFetcher) GetRaceWeek(ctx context.Context) (*RaceWeek, error) {
	now := c.now()
	for _, year := range []int{now.Year(), now.Year() + 1} {
		calendar, err := c.getF1Calendar(ctx, year)
		if err != nil {
			return nil, err
		}
		if calendar == nil {
			continue
		}
		if rw := c.mapF1CalendarToRaceWeek(calendar, now); rw != nil {
			return rw, nil
		}
	}
	return nil, nil
}

// getF1Calendar returns nil, nil when the season is not published yet.
func (c *CalendarFetcher) getF1Calendar(ctx context.Context, year int) (*F1Calendar, error) {
	labeler := &otelhttp.Labeler{}
	labeler.Add(semconv.ServicePeerName(peerName))
	ctx = otelhttp.ContextWithLabeler(ctx, labeler)

	url := fmt.Sprintf("%s/%d.json", c.baseURL, year)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("building calendar request: %w", err)
	}

	r, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching calendar %d: %w", year, err)
	}
	defer r.Body.Close()

	switch r.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, nil
	default:
		return nil, fmt.Errorf("fetching calendar %d: unexpected status %d", year, r.StatusCode)
	}

	var calendar F1Calendar
	if err := json.NewDecoder(r.Body).Decode(&calendar); err != nil {
		return nil, fmt.Errorf("decoding calendar %d: %w", year, err)
	}
	return &calendar, nil
}

func (c *CalendarFetcher) mapF1CalendarToRaceWeek(calendar *F1Calendar, now time.Time) *RaceWeek {
	for _, race := range calendar.Races {
		sessions := c.raceSessions(race.Sessions, now)
		if len(sessions) > 0 {
			return &RaceWeek{
				Location: race.Location,
				Round:    int(race.Round),
				Sessions: sessions,
			}
		}
	}
	return nil
}

func (c *CalendarFetcher) raceSessions(raceSessions Sessions, now time.Time) []SessionToBeDone {
	sessions := make([]SessionToBeDone, 0, len(raceSessions))

	for name, raw := range raceSessions {
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			slog.Warn("skipping session with invalid time",
				slog.String("session", name), slog.String("time", raw), slog.Any("error", err))
			continue
		}
		if now.Before(t) {
			sessions = append(sessions, SessionToBeDone{Name: name, Time: t.In(c.loc)})
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Time.Before(sessions[j].Time)
	})

	return sessions
}
