package models

import "errors"

type Period string

const (
	Day   = "day"
	Week  = "week"
	Month = "month"
)

type PeriodParams struct {
	WindowInDays uint64
	ToStartOf    string
	ToInterval   string
}

var PeriodParamsMap = map[Period]PeriodParams{
	Day: {
		WindowInDays: 1,
		ToStartOf:    "toStartOfHour",
		ToInterval:   "toIntervalHour",
	},
	Week: {
		WindowInDays: 7,
		ToStartOf:    "toStartOfDay",
		ToInterval:   "toIntervalDay",
	},
	Month: {
		WindowInDays: 30,
		ToStartOf:    "toStartOfDay",
		ToInterval:   "toIntervalDay",
	},
}

func ParsePeriod(s string) (Period, error) {
	switch s {
	case string(Day):
		return Day, nil
	case string(Week):
		return Week, nil
	case string(Month):
		return Month, nil
	default:
		return "", errors.New("invalid period value")
	}
}
