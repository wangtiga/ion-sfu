package sfu

import (
	"io"
	"sync"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/pion/ion-sfu/pkg/buffer"
	"github.com/pion/ion-sfu/pkg/common"
	"github.com/pion/ion-sfu/pkg/config"
	"github.com/pion/rtcp"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3"
)

type UpTrack interface {
	ID() string
	RID() string

	// PayloadType gets the PayloadType of the track
	//PayloadType() webrtc.PayloadType

	// Kind gets the Kind of the track
	Kind() webrtc.RTPCodecType

	// StreamID is the group this track belongs too. This must be unique
	StreamID() string

	// SSRC gets the SSRC of the track
	SSRC() webrtc.SSRC

	// Msid gets the Msid of the track
	//Msid() string

	// Codec gets the Codec of the track
	Codec() webrtc.RTPCodecParameters
}

type GRTCTrack struct {
	mu sync.RWMutex

	id       string
	streamID string

	payloadType webrtc.PayloadType
	kind        webrtc.RTPCodecType
	ssrc        webrtc.SSRC
	codec       webrtc.RTPCodecParameters
	params      webrtc.RTPParameters
	rid         string

	log common.ILogger
}

func newGRTCTrack(kind webrtc.RTPCodecType, ssrc webrtc.SSRC, rid string) *GRTCTrack {
	return &GRTCTrack{
		kind: kind,
		ssrc: ssrc,
		rid:  rid,

		log: config.NewLogger("RTCTrack").With("ssrc", ssrc).With("kind", kind),
	}
}

// ID is the unique identifier for this Track. This should be unique for the
// stream, but doesn't have to globally unique. A common example would be 'audio' or 'video'
// and StreamID would be 'desktop' or 'webcam'
func (t *GRTCTrack) ID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.id
}

// RID gets the RTP Stream ID of this Track
// With Simulcast you will have multiple tracks with the same ID, but different RID values.
// In many cases a TrackRemote will not have an RID, so it is important to assert it is non-zero
func (t *GRTCTrack) RID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.rid
}

// PayloadType gets the PayloadType of the track
func (t *GRTCTrack) PayloadType() webrtc.PayloadType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.payloadType
}

// Kind gets the Kind of the track
func (t *GRTCTrack) Kind() webrtc.RTPCodecType {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.kind
}

// StreamID is the group this track belongs too. This must be unique
func (t *GRTCTrack) StreamID() string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.streamID
}

// SSRC gets the SSRC of the track
func (t *GRTCTrack) SSRC() webrtc.SSRC {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.ssrc
}

// Msid gets the Msid of the track
func (t *GRTCTrack) Msid() string {
	return t.StreamID() + " " + t.ID()
}

// Codec gets the Codec of the track
func (t *GRTCTrack) Codec() webrtc.RTPCodecParameters {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.codec
}

// GRTCReceiver receives a video track
type GRTCReceiver struct {
	sync.Mutex
	rtcpMu sync.RWMutex

	peerID         string
	trackID        string
	streamID       string
	kind           webrtc.RTPCodecType
	bandwidth      uint64
	lastPli        int64
	stream         string
	codec          webrtc.RTPCodecParameters
	rtcpCh         chan []rtcp.Packet
	buffers        [3]*buffer.Buffer
	upTracks       [3]UpTrack
	downTracks     [3][]*DownTrack
	nackWorker     *workerpool.WorkerPool
	isSimulcast    bool
	onCloseHandler func()

	log common.ILogger
}

// NewGRTCReceiver creates a new webrtc track receivers
func NewGRTCReceiver(track UpTrack, pid string) Receiver {
	return &GRTCReceiver{
		peerID:      pid,
		trackID:     track.ID(),
		streamID:    track.StreamID(),
		codec:       track.Codec(),
		kind:        track.Kind(),
		nackWorker:  workerpool.New(1),
		isSimulcast: len(track.RID()) > 0,

		log: config.NewLogger("RTCReceiver").With("ssrc", track.SSRC()).With("kind", track.Kind()),
	}
}

func (w *GRTCReceiver) StreamID() string {
	return w.streamID
}

func (w *GRTCReceiver) TrackID() string {
	return w.trackID
}

func (w *GRTCReceiver) SSRC(layer int) uint32 {
	if track := w.upTracks[layer]; track != nil {
		return uint32(track.SSRC())
	}
	return 0
}

func (w *GRTCReceiver) Codec() webrtc.RTPCodecParameters {
	return w.codec
}

func (w *GRTCReceiver) Kind() webrtc.RTPCodecType {
	return w.kind
}

func (w *GRTCReceiver) AddUpTrack(track UpTrack, buff *buffer.Buffer) {
	var layer int

	switch track.RID() {
	case fullResolution:
		layer = 2
	case halfResolution:
		layer = 1
	default:
		layer = 0
	}

	w.upTracks[layer] = track
	w.buffers[layer] = buff
	w.downTracks[layer] = make([]*DownTrack, 0, 10)
	go w.writeRTP(layer)
}

func (w *GRTCReceiver) AddDownTrack(track *DownTrack, bestQualityFirst bool) {
	layer := 0
	if w.isSimulcast {
		for i, t := range w.upTracks {
			if t != nil {
				layer = i
				if !bestQualityFirst {
					break
				}
			}
		}
		track.currentSpatialLayer = layer
		track.simulcast.targetSpatialLayer = layer
		track.trackType = SimulcastDownTrack
	} else {
		track.trackType = SimpleDownTrack
	}

	w.Lock()
	w.downTracks[layer] = append(w.downTracks[layer], track)
	w.Unlock()
}

func (w *GRTCReceiver) SubDownTrack(track *DownTrack, layer int) error {
	w.Lock()
	if dts := w.downTracks[layer]; dts != nil {
		w.downTracks[layer] = append(dts, track)
	} else {
		w.Unlock()
		return errNoReceiverFound
	}
	w.Unlock()
	return nil
}

// OnCloseHandler method to be called on remote tracked removed
func (w *GRTCReceiver) OnCloseHandler(fn func()) {
	w.onCloseHandler = fn
}

// DeleteDownTrack removes a DownTrack from a Receiver
func (w *GRTCReceiver) DeleteDownTrack(layer int, id string) {
	w.Lock()
	idx := -1
	for i, dt := range w.downTracks[layer] {
		if dt.peerID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		w.Unlock()
		return
	}
	w.downTracks[layer][idx] = w.downTracks[layer][len(w.downTracks[layer])-1]
	w.downTracks[layer][len(w.downTracks[layer])-1] = nil
	w.downTracks[layer] = w.downTracks[layer][:len(w.downTracks[layer])-1]
	w.Unlock()
}

func (w *GRTCReceiver) SendRTCP(p []rtcp.Packet) {
	if _, ok := p[0].(*rtcp.PictureLossIndication); ok {
		w.rtcpMu.Lock()
		defer w.rtcpMu.Unlock()
		if time.Now().UnixNano()-w.lastPli < 500e6 {
			return
		}
		w.lastPli = time.Now().UnixNano()
	}

	w.rtcpCh <- p
}

func (w *GRTCReceiver) SetRTCPCh(ch chan []rtcp.Packet) {
	w.rtcpCh = ch
}

func (w *GRTCReceiver) RetransmitPackets(track *DownTrack, packets []uint16, snOffset uint16) {
	w.nackWorker.Submit(func() {
		pktBuff := packetFactory.Get().([]byte)
		for _, sn := range packets {
			i, err := w.buffers[track.currentSpatialLayer].GetPacket(pktBuff, sn+snOffset)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}
			var pkt rtp.Packet
			if err = pkt.Unmarshal(pktBuff[:i]); err != nil {
				continue
			}
			if err = track.WriteRTP(pkt); err == io.EOF {
				break
			}
		}
		packetFactory.Put(pktBuff)
	})
}

func (w *GRTCReceiver) writeRTP(layer int) {
	defer func() {
		w.closeTracks(layer)
		w.nackWorker.Stop()
		if w.onCloseHandler != nil {
			w.onCloseHandler()
		}
	}()
	w.log.With("layer", layer).Info("WriteRTP")
	for pkt := range w.buffers[layer].PacketChan() {
		w.log.
			With("pktSSRC", pkt.SSRC).
			With("pktSeq", pkt.SequenceNumber).
			Info("ReadFrom PacketChan")
		w.Lock()
		for _, dt := range w.downTracks[layer] {
			if err := dt.WriteRTP(pkt); err == io.EOF {
				go w.DeleteDownTrack(layer, dt.id)
			}
		}
		w.Unlock()
	}
}

// closeTracks close all tracks from Receiver
func (w *GRTCReceiver) closeTracks(layer int) {
	w.Lock()
	defer w.Unlock()
	for _, dt := range w.downTracks[layer] {
		dt.Close()
	}
}
