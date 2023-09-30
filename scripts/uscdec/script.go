package uscdec

import (
	"context"
	"embed"
	"encoding/json"

	"github.com/whitewater-guide/gorge/core"
	"golang.org/x/exp/maps"
)

type optionsUSCDEC struct {
}

type scriptUSCDEC struct {
	name string
	url  string
	core.LoggingScript
}

//go:embed cache.json
var gaugesCacheJson embed.FS

func (s *scriptUSCDEC) ListGauges() (core.Gauges, error) {
	msmnt, err := s.parseList()
	if err != nil {
		return nil, err
	}

	data, err := gaugesCacheJson.ReadFile("cache.json")
	if err != nil {
		return nil, err
	}
	cached := []core.Gauge{}
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, err
	}
	cachedCodes := map[string]core.Gauge{}
	for _, cg := range cached {
		cachedCodes[cg.Code] = cg
	}

	for _, m := range msmnt {
		if _, ok := cachedCodes[m.Code]; ok {
			continue
		} else if g, err := s.parseDetails(context.Background(), m.Code); err != nil {
			s.GetLogger().Warn(err)
		} else if g != nil {
			cachedCodes[g.Code] = *g
		}
	}

	return maps.Values(cachedCodes), nil
}

func (s *scriptUSCDEC) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	if msmnts, err := s.parseList(); err != nil {
		errs <- err
	} else {
		for _, m := range msmnts {
			recv <- m
		}
	}
}
