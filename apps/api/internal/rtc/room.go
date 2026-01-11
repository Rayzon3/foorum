package rtc

import (
	"sync"

	"github.com/pion/webrtc/v3"
)

type Room struct {
	id     string
	mu     sync.RWMutex
	peers  map[string]*Peer
	tracks map[string]trackEntry
}

type trackEntry struct {
	owner string
	track *webrtc.TrackLocalStaticRTP
}

func NewRoom(id string) *Room {
	return &Room{
		id:     id,
		peers:  make(map[string]*Peer),
		tracks: make(map[string]trackEntry),
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
	for id, entry := range r.tracks {
		if entry.owner == peerID {
			delete(r.tracks, id)
		}
	}
}

func (r *Room) AddTrack(peerID string, trackID string, track *webrtc.TrackLocalStaticRTP) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tracks[trackID] = trackEntry{owner: peerID, track: track}
}

func (r *Room) Tracks() []*webrtc.TrackLocalStaticRTP {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tracks := make([]*webrtc.TrackLocalStaticRTP, 0, len(r.tracks))
	for _, entry := range r.tracks {
		tracks = append(tracks, entry.track)
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

func (r *Room) Participants() []Participant {
	r.mu.RLock()
	defer r.mu.RUnlock()
	participants := make([]Participant, 0, len(r.peers))
	for _, peer := range r.peers {
		participants = append(participants, Participant{UserID: peer.UserID(), Role: peer.Role()})
	}
	return participants
}

func (r *Room) Broadcast(msg ServerMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, peer := range r.peers {
		select {
		case peer.Send() <- msg:
		default:
		}
	}
}
