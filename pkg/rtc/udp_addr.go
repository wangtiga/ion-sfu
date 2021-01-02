package rtc

import (
	"net"
	"strconv"
)

type UDPAddr struct {
	IP   string
	Port uint32
}

func (m UDPAddr) ToString() string {
	return m.IP + ":" + strconv.FormatUint(uint64(m.Port), 10)
}

func (m UDPAddr) UDPAddr() *net.UDPAddr {
	ret := net.UDPAddr{}
	outAddr, _ := net.ResolveUDPAddr("udp", m.ToString())
	if nil != outAddr {
		ret = *outAddr
	}
	return &ret
}
