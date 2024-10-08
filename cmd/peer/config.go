package peer

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/SotaUeda/gobgp/bgptype"
)

type Config struct {
	ConfStr  string
	LocalAS  bgptype.AutonomousSystemNumber
	LocalIP  net.IP
	RemoteAS bgptype.AutonomousSystemNumber
	RemoteIP net.IP
	Mode     Mode
	Networks []*net.IPNet
}

type Mode int

const (
	Passive Mode = iota
	Active
)

func parseMode(s string) (Mode, error) {
	switch s {
	case "passive":
		return Passive, nil
	case "active":
		return Active, nil
	default:
		return 0, fmt.Errorf("string is not mode: %s", s)
	}
}

func ParseConfig(s string) (*Config, error) {
	config := strings.Split(s, " ")
	la, err := strconv.ParseUint(config[0], 10, 16)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 1st part of config, %v, as as-number and config is %v",
			config[0], s,
		)
	}
	li := net.ParseIP(config[1]).To4()
	if li == nil {
		return nil, fmt.Errorf(
			"cannot parse 2nd part of config, %v, as as-number and config is %v",
			config[1], s,
		)
	}
	ra, err := strconv.ParseUint(config[2], 10, 16)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 3rd part of config, %v, as as-number and config is %v",
			config[2], s,
		)
	}
	ri := net.ParseIP(config[3]).To4()
	if ri == nil {
		return nil, fmt.Errorf(
			"cannot parse 4th part of config, %v, as as-number and config is %v",
			config[3], s,
		)
	}
	m, err := parseMode(config[4])
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 5th part of config, %v, as as-number and config is %v",
			config[4], s,
		)
	}
	nws := []*net.IPNet{}
	if len(config) >= 6 {
		for num, nw := range config[5:] {
			_, n, err := net.ParseCIDR(nw)
			if err != nil {
				return nil, fmt.Errorf(
					"cannot parse %vth part of config, %v as network and config is %v",
					num+5, nw, s,
				)
			}
			nws = append(nws, n)
		}
	}
	c := &Config{
		ConfStr:  s,
		LocalAS:  bgptype.AutonomousSystemNumber(la),
		LocalIP:  li,
		RemoteAS: bgptype.AutonomousSystemNumber(ra),
		RemoteIP: ri,
		Mode:     Mode(m),
		Networks: nws,
	}
	return c, nil
}
