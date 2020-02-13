package catalunya

import (
	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

func (s *scriptCatalunya) fetchObservations() ([]dataSensor, error) {
	res := &catalunyaData{}
	err := core.Client.GetAsJSON(s.measurementsURL, res, nil)

	if err != nil {
		return nil, err
	}
	return res.Sensors, err
}

func (s *scriptCatalunya) parseObservations(recv chan<- *core.Measurement, errs chan<- error) {
	dataSensors, err := s.fetchObservations()
	if err != nil {
		errs <- err
		return
	}
	for _, sensor := range dataSensors {
		for _, observation := range sensor.Observations {
			var flow, level nulltype.NullFloat64
			// observation data doesn't contain any mention of it's type
			// so this worker has to be stateful and use isFlowSensor
			ifs, err := s.isFlowSensor(&sensor)
			if err != nil {
				s.GetLogger().Error(err)
				continue
			}
			if ifs {
				flow = nulltype.NullFloat64Of(observation.Value)
			} else {
				level = nulltype.NullFloat64Of(observation.Value)
			}
			recv <- &core.Measurement{
				GaugeID: core.GaugeID{
					Script: s.name,
					Code:   sensor.Sensor,
				},
				Flow:      flow,
				Level:     level,
				Timestamp: core.HTime{Time: observation.Timestamp.Time},
			}
		}
	}
}
