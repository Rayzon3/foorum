package rtc

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"

	"jabber_v3/apps/api/internal/auth"
)

type Handler struct {
	jwt    auth.JWTManager
	rooms  *RoomManager
	api    webrtc.API
	config webrtc.Configuration
}

func NewHandler(jwt auth.JWTManager, rooms *RoomManager) *Handler {
	mediaEngine := &webrtc.MediaEngine{}
	mediaEngine.RegisterDefaultCodecs()

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	return &Handler{
		jwt:   jwt,
		rooms: rooms,
		api:   *api,
		config: webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{URLs: []string{"stun:stun.l.google.com:19302"}},
			},
		},
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request, roomID string) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("rtc panic: %v", rec)
			http.Error(w, "rtc_panic", http.StatusInternalServerError)
		}
	}()

	token := tokenFromRequest(r)
	if token == "" {
		http.Error(w, "missing_token", http.StatusUnauthorized)
		return
	}

	claims, err := h.jwt.Parse(token)
	if err != nil {
		http.Error(w, "invalid_token", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("rtc upgrade error: %v", err)
		return
	}
	defer conn.Close()

	peer := NewPeer(randomID(), claims.UserID)
	room := h.rooms.Get(roomID)
	room.AddPeer(peer)
	peer.Send() <- ServerMessage{
		Type:    "participants",
		Payload: map[string]any{"participants": room.Participants()},
	}

	done := make(chan struct{})
	go h.writeLoop(conn, peer.Send(), done)

	for {
		var msg ClientMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("rtc read error: %v", err)
			}
			break
		}

		if err := h.handleClientMessage(room, peer, msg); err != nil {
			peer.Send() <- ServerMessage{Type: "error", Error: err.Error()}
		}
	}

	close(done)
	if pc := peer.PeerConnection(); pc != nil {
		_ = pc.Close()
	}
	room.RemovePeer(peer.ID())
	room.Broadcast(ServerMessage{
		Type:    "participant_left",
		Payload: map[string]any{"userId": peer.UserID()},
	})
}

func (h *Handler) handleClientMessage(room *Room, peer *Peer, msg ClientMessage) error {
	switch msg.Type {
	case "join":
		role := strings.ToLower(stringValueFromAny(msg.Payload, "role"))
		if role == "speaker" || role == "listener" {
			peer.SetRole(role)
			room.Broadcast(ServerMessage{
				Type:    "participant_joined",
				Payload: map[string]any{"participant": Participant{UserID: peer.UserID(), Role: peer.Role()}},
			})
		}
		return nil
	case "offer":
		return h.handleOffer(room, peer, msg.SDP)
	case "candidate":
		return h.handleCandidate(peer, msg.Candidate)
	default:
		return nil
	}
}

func (h *Handler) handleOffer(room *Room, peer *Peer, sdp string) error {
	if sdp == "" {
		return errors.New("missing_sdp")
	}

	pc := peer.PeerConnection()
	if pc == nil {
		created, err := h.api.NewPeerConnection(h.config)
		if err != nil {
			return err
		}
		pc = created
		peer.SetPeerConnection(pc)

		for _, track := range room.Tracks() {
			if _, err := pc.AddTrack(track); err != nil {
				log.Printf("rtc add existing track error: %v", err)
			}
		}

		pc.OnICECandidate(func(c *webrtc.ICECandidate) {
			if c == nil {
				return
			}
			candidate := c.ToJSON()
			peer.Send() <- ServerMessage{
				Type: "candidate",
				Candidate: &ICECandidate{
					Candidate:     candidate.Candidate,
					SDPMid:        stringPtrValue(candidate.SDPMid),
					SDPMLineIndex: uint16PtrValue(candidate.SDPMLineIndex),
				},
			}
		})

		pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
			if peer.Role() != "speaker" {
				return
			}

			localTrack, err := webrtc.NewTrackLocalStaticRTP(track.Codec().RTPCodecCapability, track.ID(), track.StreamID())
			if err != nil {
				return
			}
			room.AddTrack(peer.ID(), track.ID(), localTrack)

			for _, other := range room.Peers() {
				if other.ID() == peer.ID() {
					continue
				}
				if pc := other.PeerConnection(); pc != nil {
					if _, err := pc.AddTrack(localTrack); err != nil {
						log.Printf("rtc add track error: %v", err)
					}
				}
			}

			buf := make([]byte, 1500)
			for {
				n, _, readErr := track.Read(buf)
				if readErr != nil {
					return
				}
				if _, writeErr := localTrack.Write(buf[:n]); writeErr != nil {
					return
				}
			}
		})
	}

	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdp}
	if err := pc.SetRemoteDescription(offer); err != nil {
		return err
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		return err
	}
	if err := pc.SetLocalDescription(answer); err != nil {
		return err
	}

	peer.Send() <- ServerMessage{Type: "answer", SDP: answer.SDP}
	return nil
}

func (h *Handler) handleCandidate(peer *Peer, candidate *ICECandidate) error {
	if candidate == nil || candidate.Candidate == "" {
		return nil
	}
	pc := peer.PeerConnection()
	if pc == nil {
		return nil
	}
	init := webrtc.ICECandidateInit{
		Candidate: candidate.Candidate,
	}
	if candidate.SDPMid != "" {
		init.SDPMid = &candidate.SDPMid
	}
	if candidate.SDPMLineIndex != 0 {
		init.SDPMLineIndex = &candidate.SDPMLineIndex
	}
	return pc.AddICECandidate(init)
}

func (h *Handler) writeLoop(conn *websocket.Conn, send <-chan ServerMessage, done <-chan struct{}) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-send:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

func tokenFromRequest(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	token := r.URL.Query().Get("token")
	if token != "" {
		return token
	}
	return ""
}

func stringValueFromAny(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload[key]
	if !ok {
		return ""
	}
	if value, ok := raw.(string); ok {
		return value
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	var out string
	if err := json.Unmarshal(data, &out); err != nil {
		return ""
	}
	return out
}

func stringPtrValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func uint16PtrValue(value *uint16) uint16 {
	if value == nil {
		return 0
	}
	return *value
}

func randomID() string {
	return strings.ReplaceAll(time.Now().UTC().Format("20060102150405.000000000"), ".", "")
}
