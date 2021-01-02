package sfu

import (
	//"errors"
	"fmt"
	"sync"

	//"github.com/lucsky/cuid"

	log "github.com/pion/ion-log"
	//"github.com/pion/webrtc/v3"
)

// GRTCPeer represents a pair peer connection
type GRTCPeer struct {
	sync.Mutex
	id        string
	session   *Session
	provider  SessionProvider
	publisher *GRTCPublisher
	//subscriber *Subscriber

	//OnOffer                    func(*webrtc.SessionDescription)
	//OnIceCandidate             func(*webrtc.ICECandidateInit, int)
	//OnICEConnectionStateChange func(webrtc.ICEConnectionState)

	//remoteAnswerPending bool
	//negotiationPending  bool
}

// NewGRTCPeer creates a new GRTCPeer for signaling with the given SFU
func NewGRTCPeer(pid string, provider SessionProvider) *GRTCPeer {
	return &GRTCPeer{
		id:       pid,
		provider: provider,
	}
}

// Join initializes this peer for a given sessionID (takes an SDPOffer)
func (p *GRTCPeer) Join(sid string, ssrc uint32) error {
	if p.publisher != nil {
		log.Debugf("peer already exists")
		return ErrTransportExists
	}

	//pid := cuid.New()
	//p.id = pid
	var (
		cfg WebRTCTransportConfig
		err error
	)

	p.session, cfg = p.provider.GetSession(sid)

	//p.subscriber, err = NewSubscriber(pid, cfg)
	//if err != nil {
	//	return nil, fmt.Errorf("error creating transport: %v", err)
	//}
	p.publisher, err = NewGRTCPublisher(p.session, p.id, ssrc, cfg)
	if err != nil {
		return fmt.Errorf("error creating transport: %v", err)
	}

	p.session.AddGRTCPeer(p)

	log.Infof("peer %s join session %s", p.id, sid)

	return nil
}

// Close shuts down the peer connection and sends true to the done channel
func (p *GRTCPeer) Close() error {
	if p.session != nil {
		p.session.RemoveGRTCPeer(p.id)
	}
	if p.publisher != nil {
		p.publisher.Close()
	}
	//if p.subscriber != nil {
	//	if err := p.subscriber.Close(); err != nil {
	//		return err
	//	}
	//}
	return nil
}
