package rtc

import (
	"sync"

	"github.com/pion/ion-sfu/pkg/common"
	"github.com/pion/ion-sfu/pkg/config"
	"github.com/pion/ion-sfu/pkg/mux"
)

type packetMux struct {
	lock      sync.RWMutex
	endpoints map[*packetStream]MatchFunc

	log common.ILogger
}

func newPacketMux() *packetMux {
	m := &packetMux{}
	m.endpoints = map[*packetStream]MatchFunc{}
	m.log = config.NewLogger("packetMux")
	return m
}

func (m *packetMux) dispatch(pkt *TransPacket) error {
	var endpoint *packetStream
	m.lock.Lock()
	for e, f := range m.endpoints {
		if f(pkt) {
			endpoint = e
			break
		}
	}
	m.lock.Unlock()

	if endpoint == nil {
		if len(pkt.data) > 0 {
			m.log.Warnf("Warning: mux: no endpoint for packet starting with %d\n", pkt.data[0])
		} else {
			m.log.Warnf("Warning: mux: no endpoint for zero length packet")
		}
		return nil
	}

	err := endpoint.writePkt(pkt)
	if err != nil {
		return err
	}

	return nil
}

func (m *packetMux) NewEndpoint(f MatchFunc) *packetStream {
	const KMaxTransPacket = 1024
	e := &packetStream{
		mux:   m,
		chPkt: make(chan *TransPacket, KMaxTransPacket),
	}

	m.lock.Lock()
	m.endpoints[e] = f
	m.lock.Unlock()

	return e
}

func (m *packetMux) RemoveEndpoint(e *packetStream) {
	m.lock.Lock()
	delete(m.endpoints, e)
	m.lock.Unlock()
}

type MatchFunc func(pkt *TransPacket) bool

func MatchRTCP(pkt *TransPacket) bool {
	return mux.MatchSRTCP(pkt.data)
}
func MatchRTP(pkt *TransPacket) bool {
	return mux.MatchSRTP(pkt.data)
}
