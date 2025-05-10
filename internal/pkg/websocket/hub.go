package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients organized by community ID
	clients map[int64]map[*Client]bool

	// Channel for inbound messages from clients
	broadcast chan *Message

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for concurrent access to clients map
	mu sync.RWMutex

	// Mutex for message listeners
	listenersMu sync.RWMutex
	
	// Message listeners
	messageListeners []chan *Message

	// Logger for Hub operations
	logger zerolog.Logger
}

// Message represents a message sent over WebSocket
type Message struct {
	// Type of message: "text", "file"
	Type string `json:"type"`

	// Community this message belongs to
	CommunityID int64 `json:"communityId"`

	// User who sent the message
	SenderID int64 `json:"senderId"`

	// Message content
	Content string `json:"content"`

	// Link to file if this is a file message
	FileURL string `json:"fileUrl,omitempty"`

	// File ID if this is a file message
	FileID int64 `json:"fileId,omitempty"`

	// Timestamp when the message was sent
	Timestamp time.Time `json:"timestamp"`

	// Message ID from the database
	ID int64 `json:"id,omitempty"`
}

// NewHub creates a new Hub instance
func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		broadcast:       make(chan *Message),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		clients:         make(map[int64]map[*Client]bool),
		messageListeners: []chan *Message{},
		logger:          logger,
	}
}

// Run starts the hub, handling client registrations, broadcasts, etc.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	communityID := client.communityID
	if _, ok := h.clients[communityID]; !ok {
		h.clients[communityID] = make(map[*Client]bool)
	}
	h.clients[communityID][client] = true

	h.logger.Info().
		Int64("communityID", communityID).
		Int64("userID", client.userID).
		Str("addr", client.conn.RemoteAddr().String()).
		Msg("Client registered")
}

// unregisterClient unregisters a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	communityID := client.communityID
	if _, ok := h.clients[communityID]; ok {
		if _, ok := h.clients[communityID][client]; ok {
			delete(h.clients[communityID], client)
			close(client.send)

			// If no more clients in this community, clean up
			if len(h.clients[communityID]) == 0 {
				delete(h.clients, communityID)
			}

			h.logger.Info().
				Int64("communityID", communityID).
				Int64("userID", client.userID).
				Str("addr", client.conn.RemoteAddr().String()).
				Msg("Client unregistered")
		}
	}
}

// broadcastMessage broadcasts a message to all clients in a specific community
func (h *Hub) broadcastMessage(message *Message) {
	// First, notify message listeners
	h.notifyMessageListeners(message)
	
	// Then broadcast to clients
	h.mu.RLock()
	defer h.mu.RUnlock()

	communityID := message.CommunityID
	clients, ok := h.clients[communityID]
	if !ok {
		h.logger.Debug().
			Int64("communityID", communityID).
			Msg("No clients in community for broadcast")
		return
	}

	// Serialize message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error().
			Err(err).
			Int64("communityID", communityID).
			Msg("Failed to marshal message for broadcast")
		return
	}

	// Send to all clients in this community
	for client := range clients {
		select {
		case client.send <- data:
			// Message sent successfully
		default:
			// Client's send buffer is full, they might be slow or disconnected
			// Unregister and close their connection
			h.mu.RUnlock()
			h.unregister <- client
			h.mu.RLock()
		}
	}

	h.logger.Debug().
		Int64("communityID", communityID).
		Int("clientCount", len(clients)).
		Msg("Message broadcasted to community")
}

// notifyMessageListeners sends a message to all registered message listeners
func (h *Hub) notifyMessageListeners(message *Message) {
	h.listenersMu.RLock()
	defer h.listenersMu.RUnlock()

	// Make a copy of the message for each listener
	for _, listener := range h.messageListeners {
		// Use non-blocking send to avoid blocking on slow listeners
		select {
		case listener <- message:
			// Message sent successfully
		default:
			// Listener's channel is full or slow, skip it
			h.logger.Warn().Msg("Skipped slow message listener")
		}
	}
}

// BroadcastToCommunity sends a message to all connected clients in a community
func (h *Hub) BroadcastToCommunity(message *Message) {
	h.broadcast <- message
}

// GetClientsCount returns the number of connected clients for a community
func (h *Hub) GetClientsCount(communityID int64) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.clients[communityID]; ok {
		return len(clients)
	}
	return 0
}

// AddMessageListener registers a channel to receive all messages 
func (h *Hub) AddMessageListener(listener chan *Message) {
	h.listenersMu.Lock()
	defer h.listenersMu.Unlock()
	
	h.messageListeners = append(h.messageListeners, listener)
	h.logger.Info().Msg("Added new message listener")
}

// RemoveMessageListener removes a listener from the hub
func (h *Hub) RemoveMessageListener(listener chan *Message) {
	h.listenersMu.Lock()
	defer h.listenersMu.Unlock()
	
	for i, l := range h.messageListeners {
		if l == listener {
			// Remove listener by replacing it with the last one and truncating
			h.messageListeners[i] = h.messageListeners[len(h.messageListeners)-1]
			h.messageListeners = h.messageListeners[:len(h.messageListeners)-1]
			h.logger.Info().Msg("Removed message listener")
			break
		}
	}
}