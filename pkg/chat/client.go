package chat

import (
	"sync"

	"github.com/coder/websocket"

	"github.com/phoenix-industries/omega-server/pkg/database/models"
)

type MessageType string

const (
	MessageTypeChat    MessageType = "chat"
	MessageTypeSystem  MessageType = "system"
	MessageTypeHistory MessageType = "history"
)

type Message struct {
	Type  MessageType         `json:"type"`
	Error string              `json:"error,omitempty"`
	Data  *models.ChatMessage `json:"data,omitempty"`
}

type Client struct {
	UserID  string
	Message chan *Message
	mu      sync.Mutex
	conn    *websocket.Conn
}

func NewClient(conn *websocket.Conn, userID string, messageBuffer int) *Client {
	return &Client{
		UserID:  userID,
		Message: make(chan *Message, messageBuffer),
		conn:    conn,
	}
}

func (c *Client) closeSlow() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close(websocket.StatusPolicyViolation, "connection too slow to keep up with messages")
	}
	c.conn = nil
}
