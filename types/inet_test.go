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
}
