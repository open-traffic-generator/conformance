package plot

import "time"

type Duration struct {
	ApiName  string        `json:"api_name"`
	Duration time.Duration `json:"duration"`
	Time     time.Time     `json:"time"`
}
