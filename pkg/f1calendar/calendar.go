package f1calendar

type F1Calendar struct {
	Races []Race `json:"races"`
}

type Race struct {
	Name      string   `json:"name"`
	Location  string   `json:"location"`
	Latitude  float64  `json:"latitude"`
	Longitude float64  `json:"longitude"`
	Round     int64    `json:"round"`
	Slug      string   `json:"slug"`
	LocaleKey string   `json:"localeKey"`
	Sessions  Sessions `json:"sessions"`
	Affiliate *string  `json:"affiliate,omitempty"`
	Tbc       *bool    `json:"tbc,omitempty"`
}

type Sessions map[string]string
