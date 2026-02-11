package hub

import (
	"sync"

	ws "github.com/gorilla/websocket"
)

type Client struct {
	Hub      *RoomHub
	Conn     *ws.Conn
	Send     chan []byte
	UserUID  string          // ⭐ uint → string
	Nickname string
	Rooms    map[string]bool // ⭐ uint → string (roomID)
	mu       sync.RWMutex
}

func NewClient(hub *RoomHub, conn *ws.Conn, userUID string, nickname string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserUID:  userUID,
		Nickname: nickname,
		Rooms:    make(map[string]bool),
	}
}

func (c *Client) IsInRoom(roomID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rooms[roomID]
}

func (c *Client) AddRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Rooms[roomID] = true
}

func (c *Client) RemoveRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Rooms, roomID)
}
