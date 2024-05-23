package disk

import (
	"encoding/json"
)

type IOStat struct {
	Name    string
	Tps     float64
	RdSpeed float64
	WrSpeed float64
}

type IOStatMap map[string]*IOStat

func (d IOStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

type IOCollector interface {
	Run() (<-chan IOStat, error)
	Get() IOStat
}

type UsageStat struct {
	Device                string
	Mountpoint            string
	Type                  string
	UsagePercent          float64
	Usage                 float64
	INodeCount            float64
	INodeAvailablePercent float64
}

type UsageStatMap map[string]*UsageStat

func (d UsageStat) String() string {
	s, _ := json.Marshal(d)
	return string(s)
}

type UsageCollector interface {
	Run() (<-chan UsageStatMap, error)
	Get() *UsageStat
}
