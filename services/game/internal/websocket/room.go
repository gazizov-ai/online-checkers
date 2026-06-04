package websocket

import "sync"

type GameRoom struct {
	GameID string

	mu      sync.RWMutex
	clients map[string]*Client
}

func NewGameRoom(gameID string) *GameRoom {
	return &GameRoom{
		GameID:  gameID,
		clients: make(map[string]*Client),
	}
}

func (r *GameRoom) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.UserID] = client
}

func (r *GameRoom) RemoveClient(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, userID)
}

func (r *GameRoom) CloseAll() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, client := range r.clients {
		_ = client.Close()
		delete(r.clients, userID)
	}
}

func (r *GameRoom) Broadcast(message OutgoingMessage) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, client := range r.clients {
		_ = client.Send(message)
	}
}
