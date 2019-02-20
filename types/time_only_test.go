package types

import (
	"fmt"
	"testing"
)

func TestParseTime(t *testing.T) {
	cases := []struct {
		value         uint32
		str           string
		expectError bool
	}{
		{0, "00:00:00", false},
		{86399, "23:59:59", false},
		{7158, "01:59:18", false},
		{86400, "", true},
		{100000, "", true},
	}

	for _, v := range cases {
		t.Run(fmt.Sprintf("Check time %s", v.str), func(t *testing.T) {
			str, err := ParseTime(v.value)
			if err != nil && !v.expectError {
				t.Errorf("Got unexpected error: %s", err)
			}
			if v.expectError && err == nil {
				t.Errorf("Expected error but didn't get any")
			}
			if str != v.str {
				t.Errorf("Expected value: %s, got %s", v.str, str)
			}
		})
	}
}

func TestTimeOnlyByString(t *testing.T) {
	cases := []struct {
		value       uint32
		str         string
		expectError bool
	}{
		{0, "0000-00-00T00:00:00Z", false},
		{86399, "0000-00-00T23:59:59Z", false},
		{7158, "0000-00-00T01:59:18Z", false},
		{0, "0000-00-00T24:00:00Z", true},
		{0, "0000-00-00T20:98:00Z", true},
		{0, "20:03:60Z", true},
		{0, "53:03:32Z", true},
	}

	for _, v := range cases {
		t.Run(fmt.Sprintf("Check time %s", v.str), func(t *testing.T) {
			timeOnly, err := TimeOnlyByString(v.str)
			if err != nil && !v.expectError {
				t.Errorf("Got unexpected error: %s", err)
			}
			if v.expectError && err == nil {
				t.Errorf("Expected error but didn't get any")
			}
			if !v.expectError && timeOnly.Value != v.value {
				t.Errorf("Expected value: %d, got %d", v.value, timeOnly.Value)
			}
		})
	}
}
