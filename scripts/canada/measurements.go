package canada

import (
	"errors"
	"strconv"
	"time"

	"github.com/mattn/go-nulltype"
	"github.com/whitewater-guide/gorge/core"
)

func parseDate(date string) (time.Time, error) {
	// https://stackoverflow.com/questions/27216457/best-way-of-parsing-date-and-time-in-golang
	// 2020-01-17T00:00:00-03:30
	if len(date) != 25 {
		return time.Time{}, errors.New("timestamp should be 25 symbols long")
	}
	tzHour := (int(date[20])-'0')*10 + int(date[21]) - '0'
	tzMinute := (int(date[23])-'0')*10 + int(date[24]) - '0'

	year := (((int(date[0])-'0')*10+int(date[1])-'0')*10+int(date[2])-'0')*10 + int(date[3]) - '0'
	month := time.Month((int(date[5])-'0')*10 + int(date[6]) - '0')
	day := (int(date[8])-'0')*10 + int(date[9]) - '0'
	hour := (int(date[11])-'0')*10 + int(date[12]) - '0'
	minute := (int(date[14])-'0')*10 + int(date[15]) - '0'
	second := (int(date[17])-'0')*10 + int(date[18]) - '0'

	return time.Date(year, month, day, hour, minute, second, 0, time.UTC).Add(time.Duration((tzMinute + 60*tzHour)) * time.Minute), nil
}

func getPairedGauge(code string) string {
	char := code[len(code)-3]
	if char == '0' {
		return code[:len(code)-3] + "X" + code[len(code)-2:]
	} else if char == 'X' {
		return code[:len(code)-3] + "0" + code[len(code)-2:]
	}
	return code
}

func (s *scriptCanada) measurementFromRow(line []string) (*core.Measurement, error) {
	var level, flow nulltype.NullFloat64
	if line[2] != "" {
		l, err := strconv.ParseFloat(line[2], 64)
		if err != nil {
			return nil, err
		}
		level = nulltype.NullFloat64Of(l)
	}
	if line[6] != "" {
		f, err := strconv.ParseFloat(line[6], 64)
		if err != nil {
			return nil, err
		}
		flow = nulltype.NullFloat64Of(f)
	}
	// 2020-01-17T00:00:00-03:30
	// t, err := time.Parse("2006-01-02T15:04:05-07:00", line[1])
	t, err := parseDate(line[1])
	if err != nil {
		return nil, err
	}
	return &core.Measurement{
		GaugeID: core.GaugeID{
			Script: s.name,
			Code:   line[0],
		},
		Level:     level,
		Flow:      flow,
		Timestamp: core.HTime{Time: t},
	}, nil
}
