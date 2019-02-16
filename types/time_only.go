package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

const (
	maxSeconds      uint64 = 86400
	secondsInHour   uint64 = 3600
	secondsInMinute uint64 = 60
)

var validators = []struct{
	regexp *regexp.Regexp
	timeExtractor func(string) string
}{
	{
		regexp: regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$`),
		timeExtractor: func(s string) string {
			return s[11 : 19]
		},
	},
	{
		regexp: regexp.MustCompile(`^[0-9]{2}:[0-9]{2}:[0-9]{2}$`),
		timeExtractor: func(s string) string {
			return s
		},
	},

}

func ParseTime(value uint64) (string, error) {
	t := &TimeOnly{Value: value}
	str, err := t.StringRepresentation()
	if err != nil {
		return "", err
	}
	return str, nil
}

func ParseValue(value string) (*TimeOnly, error) {
	return TimeOnlyByString(value)
}

func (t *TimeOnly) StringRepresentation() (string, error) {
	if !t.Valid() {
		return "", errors.New(fmt.Sprintf("The time exceeds %d (max value): %d", maxSeconds, t.Value))
	}

	h := t.Value / secondsInHour
	m := (t.Value - h*secondsInHour) / secondsInMinute
	s := (t.Value - h*secondsInHour - m*secondsInMinute)
	return fmt.Sprintf("%s:%s:%s", uintToStringWithLeadingZero(h), uintToStringWithLeadingZero(m), uintToStringWithLeadingZero(s)), nil
}

func TimeOnlyByString(s string) (*TimeOnly, error) {
	t := ""
	for _, v := range validators {
		if v.regexp.MatchString(s) {
			t = v.timeExtractor(s)
			break
		}
	}
	if t == "" {
		return nil, errors.New(fmt.Sprintf("Provided string %s does not represent time or simple time", t))
	}
	return getTimeOnly(t)
}

func getTimeOnly(t string) (*TimeOnly, error)   {
	h, _ := strconv.Atoi(t[0:2])
	if h > 23 || h < 0 {return nil, errors.New(fmt.Sprintf("Hours value outside expected range: %d", h))}
	m, _ := strconv.Atoi(t[3:5])
	if m > 59 || m < 0 {return nil, errors.New(fmt.Sprintf("Minutes value outside expected range: %d", m))}
	s, _ := strconv.Atoi(t[6:8])
	if s > 59 || s < 0 {return nil, errors.New(fmt.Sprintf("Seconds value outside expected range: %d", h))}
	result := uint64(h)*secondsInHour + uint64(m)*secondsInMinute + uint64(s)
	time := &TimeOnly{Value: result}
	if !time.Valid() {
		return nil, errors.New(fmt.Sprintf("The time exceeds %d (max value): %d", maxSeconds, result))
	}
	return time, nil
}

func uintToStringWithLeadingZero(t uint64) string {
	out := strconv.Itoa(int(t))
	if len(out) == 1 {
		out = "0" + out
	}
	return out
}

func (t *TimeOnly) Valid() bool {
	return t.Value < maxSeconds
}
