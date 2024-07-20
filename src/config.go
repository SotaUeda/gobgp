package peer

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Config struct {
	ConfStr  string
	LocalAS  AutonomousSystemNumber
	LocalIP  net.IP
	RemoteAS AutonomousSystemNumber
	RemoteIP net.IP
	Mode     Mode
}

type Mode int

const (
	Passive Mode = iota
	Active
)

func Str2Config(s string) (*Config, error) {
	config := strings.Split(s, " ")
	la, err := strconv.ParseUint(config[0], 10, 16)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 1st part of config, %v, as as-number and config is %v",
			config[0], s,
		)
	}
	li := net.ParseIP(config[1])
	if li == nil {
		return nil, fmt.Errorf(
			"cannot parse 2nd part of config, %v, as as-number and config is %v",
			config[1], s,
		)
	}
	ra, err := strconv.ParseUint(config[0], 10, 16)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 3rd part of config, %v, as as-number and config is %v",
			config[2], s,
		)
	}
	ri := net.ParseIP(config[1])
	if ri == nil {
		return nil, fmt.Errorf(
			"cannot parse 4th part of config, %v, as as-number and config is %v",
			config[3], s,
		)
	}
	m, err := strconv.ParseInt(config[4], 10, 64)
	if err != nil {
		return nil, fmt.Errorf(
			"cannot parse 5th part of config, %v, as as-number and config is %v",
			config[4], s,
		)
	}
	c := &Config{
		ConfStr:  s,
		LocalAS:  AutonomousSystemNumber(la),
		LocalIP:  li,
		RemoteAS: AutonomousSystemNumber(ra),
		RemoteIP: ri,
		Mode:     Mode(m),
	}
	return c, nil
}
