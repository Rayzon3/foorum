package rtc

type ClientMessage struct {
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	SDP       string         `json:"sdp,omitempty"`
	Candidate *ICECandidate  `json:"candidate,omitempty"`
}

type ServerMessage struct {
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
	SDP       string         `json:"sdp,omitempty"`
	Candidate *ICECandidate  `json:"candidate,omitempty"`
	Error     string         `json:"error,omitempty"`
}

type ICECandidate struct {
	Candidate     string `json:"candidate"`
	SDPMid        string `json:"sdpMid,omitempty"`
	SDPMLineIndex uint16 `json:"sdpMLineIndex,omitempty"`
}
