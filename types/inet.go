package types

import (
	"database/sql/driver"
	"errors"
	"net"
)

// Inet is a special scannable type for an IP/Netmask
type Inet struct {
	*net.IPNet
}

// Value implements the Value part of the sql scannable interface
func (i Inet) Value() (driver.Value, error) {
	if i.IPNet == nil {
		return nil, nil
	}
	return []byte(i.String()), nil
}

// Scan implements the scan part of the sql scannable interface
func (i *Inet) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var strdat string
	bytes, ok := value.([]byte)
	if !ok {
		if strdat, ok = value.(string); !ok {
			return errors.New("Could not cast value in Inet.Scan as []byte or string")
		}
	} else {
		strdat = string(bytes)
	}
	ip, cidr, err := net.ParseCIDR(strdat)
	if err != nil {
		return err
	}
	i.IPNet = &net.IPNet{IP: ip, Mask: cidr.Mask}
	return nil
}

// ParseInet will return the Inet address/netmask represented in the input string
func ParseInet(addr string) (*Inet, error) {
	ip, cidr, err := net.ParseCIDR(addr)
	if err != nil {
		return nil, err
	}
	return &Inet{&net.IPNet{IP: ip, Mask: cidr.Mask}}, err
}
