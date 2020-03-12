package types

import (
	"net"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func TestInetParse(t *testing.T) {
	inet, err := ParseInet("1.2.3.4/32")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("1.2.3.4")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(32, 32)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
	// ------
	inet, err = ParseInet("1.2.3.4/24")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("1.2.3.4")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(24, 32)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
	// ------
	inet, err = ParseInet("fe80:3::1ff:fe23:4567:890a/48")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("fe80:3::1ff:fe23:4567:890a")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(48, 128)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
	// ------
	inet, err = ParseInet("1.2.3.4")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("1.2.3.4")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(32, 32)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
	// ------
	inet, err = ParseInet("fe80:3::1ff:fe23:4567:890a")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("fe80:3::1ff:fe23:4567:890a")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(128, 128)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
	// ------
	inet, err = ParseInet("[2000::1]")
	if err != nil {
		t.Error(err)
	}
	if !inet.IP.Equal(net.ParseIP("2000::1")) || !reflect.DeepEqual(inet.Mask, net.CIDRMask(128, 128)) {
		t.Errorf("Did not get expected value, got %+v", *inet)
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{"v4", "192.168.1.1", "192.168.1.1"},
		{"v4 with single host mask", "192.168.1.1/32", "192.168.1.1"},
		{"v4 with different mask", "192.168.1.1/24", "192.168.1.1/24"},
		{"v6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "2001:db8:85a3::8a2e:370:7334"},
		{"v6 with single host mask", "2001:0db8:85a3:0000:0000:8a2e:0370:7334/128", "2001:db8:85a3::8a2e:370:7334"},
		{"v6 with different", "2001:0db8:85a3:0000:0000:8a2e:0370:7334/64", "2001:db8:85a3::8a2e:370:7334/64"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if inet, err := ParseInet(tc.input); err != nil {
				t.Errorf("failed to parse Inet value %s: %v", tc.input, err)
			} else if got, want := inet.String(), tc.want; want != got {
				t.Errorf("got %s; want %s", got, want)
			}
		})
	}
}
