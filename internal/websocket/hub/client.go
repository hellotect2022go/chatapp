package hub

import (
	"sync"

	ws "github.com/gorilla/websocket"
)

type Client struct {
	Hub      *RoomHub
	Conn     *ws.Conn
	Send     chan []byte
	UserID   uint
	Nickname string
	Rooms    map[uint]bool // 해당 클라이언트가 참여한 방의 목록들
	mu       sync.RWMutex
}

func NewClient(hub *RoomHub, conn *ws.Conn, userID uint, nickname string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   userID,
		Nickname: nickname,
		Rooms:    make(map[uint]bool),
	}
}

func (c *Client) IsInRoom(roomID uint) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rooms[roomID]
}

func (c *Client) AddRoom(roomID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Rooms[roomID] = true
}

func (c *Client) RemoveRoom(roomID uint) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Rooms, roomID)
}
