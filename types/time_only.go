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

var validTime = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$`)
var validSimpleTime = regexp.MustCompile(`^[0-9]{2}:[0-9]{2}:[0-9]{2}$`)

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

func TimeOnlyBySimpleString(t string) (*TimeOnly, error) {
	if !validTime.MatchString(t) {

		return nil, errors.New(fmt.Sprintf("Provided string %s does not represent simple time", t))
	}
	return getTimeOnly(t)
}

func TimeOnlyByString(t string) (*TimeOnly, error) {
	if !validTime.MatchString(t) {
		return nil, errors.New(fmt.Sprintf("Provided string %s does not represent time", t))
	}
	return getTimeOnly(t[11 : 19])
}

func getTimeOnly(t string) (*TimeOnly, error)   {
	h, _ := strconv.Atoi(t[11:13])
	m, _ := strconv.Atoi(t[14:16])
	s, _ := strconv.Atoi(t[17:19])
	fmt.Println(t[11:12])
	fmt.Println(t[14:15])
	fmt.Println(t[17:18])
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
