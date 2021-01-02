package sfu

import (
	"net"
	"sync"
	//"sync/atomic"

	log "github.com/pion/ion-log"
	"github.com/pion/ion-sfu/pkg/buffer"
	"github.com/pion/rtcp"
	"github.com/pion/transport/packetio"
	"github.com/pion/webrtc/v3"
)

type GRTCConnection struct {
	conn    net.UDPConn
	dstAddr string

	mu             sync.RWMutex
	onTrackHandler func(*webrtc.TrackRemote, *webrtc.RTPReceiver)
}

func (pc *GRTCConnection) OnTrack(f func(*webrtc.TrackRemote, *webrtc.RTPReceiver)) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.onTrackHandler = f
}

func (pc *GRTCConnection) WriteRTCP(pkts []rtcp.Packet) error {
	return nil
}

func (pc *GRTCConnection) Close() error {
	return nil
}

type GRTCPublisher struct {
	id string
	pc *GRTCConnection

	router  Router
	session *Session

	onTrackHandler func(*webrtc.TrackRemote, *webrtc.RTPReceiver)

	closeOnce sync.Once
}

// NewGRTCPublisher creates a new GRTCPublisher
func NewGRTCPublisher(session *Session, id string, ssrc uint32, cfg WebRTCTransportConfig) (*GRTCPublisher, error) {

	pc := &GRTCConnection{} // TODO
	p := &GRTCPublisher{
		id:      id,
		pc:      pc,
		session: session,
		router:  newGRTCRouter(pc, id, cfg.router),
	}

	//pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
	//	log.Debugf("Peer %s got remote track id: %s mediaSSRC: %d rid :%s streamID: %s", p.id, track.ID(), track.SSRC(), track.RID(), track.StreamID())
	//	if r, pub := p.router.AddReceiver(receiver, track); pub {
	//		p.session.Publish(p.router, r)
	//	}
	//})

	// TODO for test
	track := newGRTCTrack(
		webrtc.RTPCodecTypeVideo,
		webrtc.SSRC(ssrc),
		"",
	)

	buff := bufferFactory.GetOrNew(packetio.RTPBufferPacket, uint32(track.SSRC())).(*buffer.Buffer)
	rtcpReader := bufferFactory.GetOrNew(packetio.RTCPBufferPacket, uint32(track.SSRC())).(*buffer.RTCPReader)
	if nil != buff {
	}
	if nil != rtcpReader {
	}

	videoRTCPFeedback := []webrtc.RTCPFeedback{{"goog-remb", ""}, {"ccm", "fir"}, {"nack", ""}, {"nack", "pli"}}
	opusCodec := webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeOpus, 48000, 2, "minptime=10;useinbandfec=1", nil},
		PayloadType:        111,
	}
	h264Codec := webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{webrtc.MimeTypeH264, 90000, 0, "level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f", videoRTCPFeedback},
		PayloadType:        125,
	}

	rtpParams := webrtc.RTPParameters{
		Codecs: []webrtc.RTPCodecParameters{opusCodec},
	}

	rtpParams = webrtc.RTPParameters{
		Codecs: []webrtc.RTPCodecParameters{h264Codec},
	}

	if r, pub := p.router.AddGRTCReceiver(rtpParams, track); pub {
		p.session.Publish(p.router, r)
	}

	return p, nil
}

// GetRouter returns router with mediaSSRC
func (p *GRTCPublisher) GetRouter() Router {
	return p.router
}

// Close peer
func (p *GRTCPublisher) Close() {
	p.closeOnce.Do(func() {
		p.router.Stop()
		if err := p.pc.Close(); err != nil {
			log.Errorf("webrtc transport close err: %v", err)
		}
	})
}
