package rtc

import "sync"

type RoomManager struct {
	mu    sync.Mutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{rooms: make(map[string]*Room)}
}

func (m *RoomManager) Get(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if room, ok := m.rooms[roomID]; ok {
		return room
	}
	room := NewRoom(roomID)
	m.rooms[roomID] = room
	return room
}
