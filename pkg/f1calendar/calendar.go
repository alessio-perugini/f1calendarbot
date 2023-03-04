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

type Sessions struct {
	Fp1        string  `json:"fp1"`
	Fp2        string  `json:"fp2"`
	Fp3        *string `json:"fp3,omitempty"`
	Qualifying string  `json:"qualifying"`
	Gp         string  `json:"gp"`
	Sprint     *string `json:"sprint,omitempty"`
}
