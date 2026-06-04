package websocket

import "sync"

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*GameRoom
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*GameRoom),
	}
}

func (m *RoomManager) GetOrCreateRoom(gameID string) *GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, ok := m.rooms[gameID]
	if ok {
		return room
	}

	room = NewGameRoom(gameID)
	m.rooms[gameID] = room

	return room
}

func (m *RoomManager) DeleteRoom(gameID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, gameID)
}
