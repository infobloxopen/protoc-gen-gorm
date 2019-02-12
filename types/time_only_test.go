package types

import (
	"fmt"
	"testing"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		value         uint64
		str           string
		expectedError error
	}{
		{0, "00:00:00", nil},
		{86399, "23:59:59", nil},
		{7158, "01:59:18", nil},
	}

	for _, v := range cases {
		t.Run(fmt.Sprintf("Check time %s", v.str), func(t *testing.T) {
			str, err := ParseTime(v.value)
			if err != nil && v.expectedError == nil{
				t.Fail()
			}
			if str != v.str {
				t.Errorf("Expected value: %s, got %s", v.str, str)
			}
		})
	}
}


func TestParseValue(t *testing.T) {
	cases := []struct {
		value         uint64
		str           string
		expectedError error
	}{
		{0, "0000-00-00T00:00:00Z", nil},
		{86399, "0000-00-00T23:59:59Z", nil},
		{7158, "0000-00-00T01:59:18Z", nil},
	}

	for _, v := range cases {
		t.Run(fmt.Sprintf("Check time %s", v.str), func(t *testing.T) {
			timeOnly, err := ParseValue(v.str)
			if err != nil && v.expectedError == nil{
				t.Fail()
			}
			if timeOnly.Value != v.value {
				t.Errorf("Expected value: %d, got %d", v.value, timeOnly.Value)
			}
		})
	}
}
