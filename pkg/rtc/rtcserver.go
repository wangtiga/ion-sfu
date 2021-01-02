package rtc

import (
	"fmt"
	"github.com/pion/ion-sfu/pkg/buffer"
	"github.com/pion/ion-sfu/pkg/common"
	"github.com/pion/ion-sfu/pkg/config"
	"github.com/pion/transport/packetio"
)

const (
	receiveMTU = 1460 // Equal to UDP MTU

	//rtpOutboundMTU          = 1200
	trackDefaultIDLength    = 16
	trackDefaultLabelLength = 16
)

type RtcServer struct {
	inConnMaps    *InConnMap
	bufferFactory *buffer.Factory

	log common.ILogger
}

func NewRtcServer(b *buffer.Factory) *RtcServer {
	s := &RtcServer{}
	s.log = config.NewLogger("rtcserver")
	s.bufferFactory = b
	s.inConnMaps = NewInConnMap()
	return s
}

func (s *RtcServer) SetConfig(start, end int) error {
	rcvBuf := 100 * 1024 * 1024
	sndBuf := 100 * 1024 * 1024

	// TODO reset resource
	for i := start; i <= end; i++ {
		addr := UDPAddr{}
		addr.Port = uint32(i)
		if err := s.AddListen(addr, rcvBuf, sndBuf); nil != err {
			return err
		}
	}

	s.log.Infof("SetConfig  start=%v end=%v", start, end)
	return nil
}

func (s *RtcServer) handleRTP(pkt *TransPacket) error {
	buff := s.bufferFactory.GetOrNew(packetio.RTPBufferPacket, uint32(pkt.SSRC()))
	if nil == buff {
		return nil
	}

	wn, err := buff.Write(pkt.data)
	if nil != err {
		return err
	}
	if wn != len(pkt.data) {
		return fmt.Errorf("FullBuffer len(data)=%v wn=%v", len(pkt.data), wn)
	}
	return nil
}

func (s *RtcServer) handleRTCP(pkt *TransPacket) error {
	rtcpReader := s.bufferFactory.GetOrNew(packetio.RTCPBufferPacket, uint32(pkt.SSRC()))

	wn, err := rtcpReader.Write(pkt.data)
	if nil != err {
		return err
	}
	if wn != len(pkt.data) {
		return fmt.Errorf("FullBuffer len(data)=%v wn=%v", len(pkt.data), wn)
	}

	return nil
}

func (s *RtcServer) AddListen(addr UDPAddr, rcvBuf, sndBuf int) error {
	c, err := s.inConnMaps.NewInConn(addr, rcvBuf, sndBuf)
	if nil != err {
		return err
	}

	go func() {
		s.log.Info("handleRTP start")
		for {
			pkt := c.rtpEndpoint.ReadPkt()
			if nil == pkt {
				s.log.Info("handleRTP end")
				return
			}
			err := s.handleRTP(pkt)
			if nil != err {
				s.log.Warnf("handleRTP err=%v", err)
			}
		}
	}()
	go func() {
		s.log.Info("handleRTCP start")
		for {
			pkt := c.rtcpEndpoint.ReadPkt()
			if nil == pkt {
				s.log.Info("handleRTCP end")
				return
			}

			err := s.handleRTCP(pkt)
			if nil != err {
				s.log.Warnf("handleRTCP err=%v", err)
			}
		}
	}()

	s.log.Infof("AddListen %v", addr)
	return nil
}
