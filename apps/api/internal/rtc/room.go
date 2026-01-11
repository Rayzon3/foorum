package rtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type Room struct {
	id     string
	mu     sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]*webrtc.TrackLocalStaticRTP
}

func NewRoom(id string) *Room {
	return &Room{
		id:     id,
		peers:  make(map[string]*Peer),
		tracks: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
}

func (r *Room) ID() string {
	return r.id
}

func (r *Room) AddPeer(peer *Peer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.peers[peer.ID()] = peer
}

func (r *Room) RemovePeer(peerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.peers, peerID)
}

func (r *Room) AddTrack(trackID string, track *webrtc.TrackLocalStaticRTP) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracks[trackID] = track
}

func (r *Room) Tracks() []*webrtc.TrackLocalStaticRTP {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tracks := make([]*webrtc.TrackLocalStaticRTP, 0, len(r.tracks))
	for _, track := range r.tracks {
		tracks = append(tracks, track)
	}
	return tracks
}

func (r *Room) Peers() []*Peer {
	r.mu.RLock()
	defer r.mu.RUnlock()
	peers := make([]*Peer, 0, len(r.peers))
	for _, peer := range r.peers {
		peers = append(peers, peer)
	}
	return peers
}
