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
	inet, err := ParseInet(strdat)
	i.IPNet = inet.IPNet
	return err
}

// ParseInet will return the Inet address/netmask represented in the input string
func ParseInet(addr string) (*Inet, error) {
	if len(addr) == 0 {
		return nil, nil
	}
	ip, cidr, err := net.ParseCIDR(addr)
	var mask net.IPMask
	if err != nil {
		ip = net.ParseIP(addr)
		if ip == nil {
			return nil, err
		}
		if v4 := ip.To4(); v4 != nil {
			mask = net.CIDRMask(32, 32)
		} else {
			mask = net.CIDRMask(128, 128)
		}
	} else {
		mask = cidr.Mask
	}
	return &Inet{&net.IPNet{IP: ip, Mask: mask}}, nil
}

func (i *Inet) String() string {
	// don't print the mask if it specifies only a single host
	if ones, bits := i.Mask.Size(); ones == bits {
		return i.IP.String()
	}
	return i.IPNet.String()
}
