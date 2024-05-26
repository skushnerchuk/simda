package loadavg

import (
	"encoding/json"
)

type AvgStat struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

func (l AvgStat) String() string {
	s, _ := json.Marshal(l)
	return string(s)
}

type LoadAverageCollector interface {
	Run() (<-chan *AvgStat, error)
	Get() (*AvgStat, error)
}
