package rtc

import (
	"fmt"
	"net"
	//"github.com/valyala/bytebufferpool"

	"github.com/pion/rtp"
)

type TransPacket struct {
	conn *net.UDPConn
	//localAddr  net.UDPAddr
	remoteAddr net.UDPAddr

	//data *bytebufferpool.ByteBuffer
	data []byte
}

func NewTransPacket() *TransPacket {
	pkt := &TransPacket{}
	pkt.data = make([]byte, receiveMTU, receiveMTU)
	return pkt
}

func (t TransPacket) SSRC() SSRCType { // TODO optimize
	h := &rtp.Header{}
	if err := h.Unmarshal(t.data); err != nil {
		// return err
	} else {
		return SSRCType(h.SSRC)
	}

	return SSRCType(1) // TODO
}

func (t TransPacket) LenData() int {
	return len(t.data)
}

func (t TransPacket) RTP() (*rtp.Packet, error) {
	packet := &rtp.Packet{}
	err := packet.Unmarshal(t.data)
	if err != nil {
		return nil, err
	}
	return packet, nil
}

func (t TransPacket) String() string {
	h := &rtp.Header{}
	if err := h.Unmarshal(t.data); err != nil {
		// return err
	} else {
	}

	localAddr := t.conn.LocalAddr()
	remoteAddr := t.remoteAddr
	return fmt.Sprintf(
		"ssrc=%v, localAddr=%v, remoteAddr=%v",
		h.SSRC,
		localAddr,
		remoteAddr,
	)
}
