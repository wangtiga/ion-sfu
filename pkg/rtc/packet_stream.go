package rtc

import ()

type packetStream struct {
	mux   *packetMux
	chPkt chan *TransPacket
}

//func (e *packetStream) Close() (err error) {
//	close(e.pktCh)
//	e.mux.RemoveEndpoint(e)
//	return nil
//}

func (e *packetStream) ReadPkt() *TransPacket {
	return <-e.chPkt
}

func (e *packetStream) writePkt(pkt *TransPacket) error {
	e.chPkt <- pkt
	return nil
}
