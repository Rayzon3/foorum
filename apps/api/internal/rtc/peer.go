package rtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type Peer struct {
	id     string
	userID string
	role   string
	mu     sync.RWMutex
	pc     *webrtc.PeerConnection
	send   chan ServerMessage
}

func NewPeer(id string, userID string) *Peer {
	return &Peer{
		id:     id,
		userID: userID,
		role:   "listener",
		send:   make(chan ServerMessage, 32),
	}
}

func (p *Peer) ID() string {
	return p.id
}

func (p *Peer) UserID() string {
	return p.userID
}

func (p *Peer) SetRole(role string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.role = role
}

func (p *Peer) Role() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.role
}

func (p *Peer) SetPeerConnection(pc *webrtc.PeerConnection) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pc = pc
}

func (p *Peer) PeerConnection() *webrtc.PeerConnection {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.pc
}

func (p *Peer) Send() chan ServerMessage {
	return p.send
}
