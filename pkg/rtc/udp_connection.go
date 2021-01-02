package rtc

import (
	"errors"
	"net"
	"sync"

	"github.com/pion/ion-sfu/pkg/common"
	"github.com/pion/ion-sfu/pkg/config"
	"github.com/pion/rtp"
)

type InConnMap struct {
	m     sync.RWMutex
	datas map[string]*udpConn

	mOut     sync.RWMutex
	datasOut map[string]*net.UDPConn
}

func NewInConnMap() *InConnMap {
	m := &InConnMap{}
	m.datas = map[string]*udpConn{}

	m.datasOut = map[string]*net.UDPConn{}
	return m
}

func (c *InConnMap) RandomConn() (*net.UDPConn, error) {
	var conn *net.UDPConn

	c.m.RLock()
	for _, item := range c.datas {
		conn = item.conn
		break
	}
	c.m.RUnlock()

	if nil != conn {
		return conn, nil
	}

	return nil, errors.New("EmptyConn")
}

func (c *InConnMap) GetOutConn(addr UDPAddr) (*net.UDPConn, error) {
	c.mOut.RLock()
	conn, ok := c.datasOut[addr.ToString()]
	c.mOut.RUnlock()

	if ok {
		return conn, nil
	}

	c.mOut.Lock()
	defer c.mOut.Unlock()

	// get conn from c.datasIn TODO multi user subscribe same user , may cause error
	conn, _ = c.RandomConn()

	if nil != conn {
		c.datasOut[addr.ToString()] = conn
		return conn, nil
	}
	return nil, errors.New("NotSupport dstAddr=" + addr.ToString())

	conn, ok = c.datasOut[addr.ToString()]
	if ok {
		return conn, nil
	}
	outAddr, err := net.ResolveUDPAddr("udp", addr.ToString())
	if err != nil {
		return nil, errors.New("failed to ResolveUDPAddr: " + err.Error())
	}
	outConn, err := net.DialUDP("udp", nil, outAddr)
	if err != nil {
		return nil, errors.New("failed to Dial: " + err.Error())
	}

	c.datasOut[addr.ToString()] = outConn
	return outConn, nil

}

func (c *InConnMap) NewInConn(addr UDPAddr, rcvBuf, sndBuf int) (*udpConn, error) {
	c.m.RLock()
	conn, ok := c.datas[addr.ToString()]
	c.m.RUnlock()

	if ok {
		return conn, nil
	}

	c.m.Lock()
	defer c.m.Unlock()
	conn, ok = c.datas[addr.ToString()]
	if ok {
		return conn, nil
	}

	inAddr, err := net.ResolveUDPAddr("udp", addr.ToString())
	if err != nil {
		return nil, errors.New("failed to ResolveUDPAddr: " + err.Error())
	}
	inConn, err := net.ListenUDP("udp", inAddr)
	if err != nil {
		return nil, errors.New("failed to listen: " + err.Error())
	}
	inConn.SetReadBuffer(rcvBuf)
	inConn.SetWriteBuffer(sndBuf)
	conn, err = NewUdpConn(inConn)
	if err != nil {
		return nil, errors.New("failed to NewPeerConnection: " + err.Error())
	}
	c.datas[addr.ToString()] = conn
	return conn, nil
}

type udpConn struct {
	conn *net.UDPConn
	log  common.ILogger
	mux  *packetMux

	rtpEndpoint  *packetStream
	rtcpEndpoint *packetStream
}

func NewUdpConn(conn *net.UDPConn) (*udpConn, error) {

	log := config.NewLogger("udpConn")
	c := &udpConn{}
	c.conn = conn
	c.log = log
	c.mux = newPacketMux()

	c.rtpEndpoint = c.mux.NewEndpoint(MatchRTP)
	c.rtcpEndpoint = c.mux.NewEndpoint(MatchRTCP)
	go c.readLoop()
	return c, nil
}

func (m *udpConn) readLoop() {
	for {
		pkt := NewTransPacket()
		pkt.conn = m.conn

		n, remoteAddr, err := m.conn.ReadFromUDP(pkt.data)
		if err != nil {
			m.log.Warnf("Warning: udpconn: read err=%v", err)
			return
		}

		pkt.remoteAddr = *remoteAddr
		pkt.data = pkt.data[:n]
		err = m.mux.dispatch(pkt)
		if err != nil {
			m.log.Warnf("Warning: udpconn: dispatch err=%v", err)
			return
		}
	}
}

func (m *udpConn) WriteRTP(header *rtp.Header, payload []byte, rAddr net.Addr) (int, error) {
	var dst []byte = nil
	dst = growBufferSize(dst, header.MarshalSize()+len(payload))
	n, err := header.MarshalTo(dst)
	if err != nil {
		return n, err
	}
	copy(dst[n:], payload)
	return m.conn.WriteTo(dst, rAddr)
}

// Grow the buffer size to the given number of bytes.
func growBufferSize(buf []byte, size int) []byte {
	if size <= cap(buf) {
		return buf[:size]
	}

	buf2 := make([]byte, size)
	copy(buf2, buf)
	return buf2
}
