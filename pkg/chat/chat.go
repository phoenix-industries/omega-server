package chat

import (
	"context"
	"log/slog"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"

	"github.com/phoenix-industries/omega-server/pkg/database"
	"github.com/phoenix-industries/omega-server/pkg/database/models"
)

type Server struct {
	db      *database.Database
	logger  *slog.Logger
	config  *Config
	mu      sync.RWMutex
	rooms   map[string]map[string]struct{} // room_id -> user_id
	clients map[string]*Client             // user_id -> client
}

func NewServer(db *database.Database, logger *slog.Logger, config *Config) *Server {
	if db == nil {
		panic("chat: db is required")
	}
	if logger == nil {
		logger = slog.Default().WithGroup("chat")
	}
	if config == nil {
		config = DefaultConfig()
	}
	return &Server{
		db:      db,
		logger:  logger,
		config:  config,
		rooms:   make(map[string]map[string]struct{}),
		clients: make(map[string]*Client),
	}
}

func (s *Server) NewClient(conn *websocket.Conn, userID string) *Client {
	return NewClient(conn, userID, s.config.MaxMessagesBuffer)
}

func (s *Server) Subscribe(c *Client) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s.addClient(ctx, c)
	defer s.removeClient(c)

	go func() {
		defer cancel()
		for {
			var msg Message
			if err := wsjson.Read(ctx, c.conn, &msg); err != nil {
				if websocket.CloseStatus(err) == websocket.StatusNormalClosure ||
					websocket.CloseStatus(err) == websocket.StatusGoingAway {
					return
				}
				s.logger.ErrorContext(ctx, "read", "error", err)
				return
			}
			switch msg.Type {
			case MessageTypeSystem:
				// TODO
			case MessageTypeChat:
				if err := models.ChatMessageInsert(ctx, s.db.Pool(), msg.Data); err != nil {
					s.logger.ErrorContext(ctx, "chat message insert", "error", err)
					continue
				}
			}
			s.sendMessage(&msg)
		}
	}()

	for {
		select {
		case msg := <-c.Message:
			if err := s.writeTimeout(ctx, c.conn, msg); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *Server) sendMessage(m *Message) {
	if m == nil || m.Data == nil {
		return
	}
	s.config.MessageLimiter.Wait(context.Background())
	s.mu.RLock()
	defer s.mu.RUnlock()
	for userID := range s.rooms[m.Data.RoomID] {
		if userID == m.Data.UserID {
			continue
		}
		if c, ok := s.clients[userID]; ok {
			select {
			case c.Message <- m:
			default:
				go c.closeSlow()
			}
		}
	}
}

func (s *Server) addClient(ctx context.Context, c *Client) {
	rooms, err := models.ChatRoomList(ctx, s.db.Pool(), c.UserID)
	if err != nil {
		s.logger.ErrorContext(ctx, "chat room list", "error", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[c.UserID] = c
	for _, room := range rooms {
		if _, ok := s.rooms[room.ID]; !ok {
			s.rooms[room.ID] = make(map[string]struct{})
		}
		s.rooms[room.ID][c.UserID] = struct{}{}
	}
}

func (s *Server) removeClient(c *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c.UserID)
	for roomID, users := range s.rooms {
		delete(users, c.UserID)
		if len(users) == 0 {
			delete(s.rooms, roomID)
		}
	}
}

func (s *Server) AddRoom(roomID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, userID := range userIDs {
		if _, ok := s.clients[userID]; !ok {
			continue
		}
		if _, ok := s.rooms[roomID]; !ok {
			s.rooms[roomID] = make(map[string]struct{})
		}
		s.rooms[roomID][userID] = struct{}{}
	}
	return nil
}

func (s *Server) writeTimeout(ctx context.Context, conn *websocket.Conn, v any) error {
	ctx, cancel := context.WithTimeout(ctx, s.config.writeTimeout)
	defer cancel()
	return wsjson.Write(ctx, conn, v)
}
