package cpu

import (
	"encoding/json"
)

type Data struct {
	User   float64
	Idle   float64
	System float64
}

func (c Data) String() string {
	s, _ := json.Marshal(c)
	return string(s)
}

type CPUCollector interface { //nolint:revive
	Get() (*Data, error)
	Run() (<-chan *Data, error)
}
