package finland

import (
	"context"
	"fmt"

	"github.com/whitewater-guide/gorge/core"
)

type optionsFinland struct{}
type scriptFinland struct {
	name string
	url  string
	core.LoggingScript
}

func (s *scriptFinland) ListGauges() (core.Gauges, error) {
	result := core.Gauges{}
	err := s.fetchList(fmt.Sprintf("%s/Paikka?$skip=0&$select=KoordLat,KoordLong,KuntaNimi,Nro,Paikka_Id,Nimi,Laitteisto_Id", s.url), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *scriptFinland) Harvest(ctx context.Context, recv chan<- *core.Measurement, errs chan<- error, codes core.StringSet, since int64) {
	defer close(recv)
	defer close(errs)
	code, err := codes.Only()
	if err != nil {
		errs <- err
		return
	}
	s.fetchMeasurement(code, recv, errs)
}
