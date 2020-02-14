package core

import "fmt"

// HarvestMode determines how the schedule for script is generated
type HarvestMode int

const (
	// AllAtOnce scripts harvest all gauges in one batch. Such sources usually provide one file with latest measurements for every gauge
	AllAtOnce HarvestMode = iota
	// OneByOne scripts harvest only one gauge at a time.
	// For such scripts a schedule will be generated so that all the gauges are uniformly harvested during period of 1 hour
	OneByOne
)

var harvestModeStr = [...]string{`"allAtOnce"`, `"oneByOne"`}

func (m HarvestMode) String() string {
	quoted := harvestModeStr[m]
	return quoted[1 : len(quoted)-1]
}

// MarshalJSON implements json.Marshaler interface
func (m *HarvestMode) MarshalJSON() ([]byte, error) {
	return []byte(harvestModeStr[*m]), nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *HarvestMode) UnmarshalJSON(bytes []byte) error {
	for i, s := range harvestModeStr {
		if s == string(bytes) {
			*m = HarvestMode(i)
			return nil
		}
	}
	return fmt.Errorf("invalid harvest mode '%s'", string(bytes))
}

// TSName is required to generate typescript enum
func (m HarvestMode) TSName() string {
	switch m {
	case AllAtOnce:
		return "AllAtOnce"
	case OneByOne:
		return "OneByOne"
	default:
		return "???"
	}
}
