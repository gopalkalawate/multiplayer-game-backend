package socket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected player
type Client struct {
	Hub      *Hub
	Game     *Game // Link to game to send inputs
	MatchID  string
	PlayerID string
	Conn     *websocket.Conn
	Send     chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to the match rooms
type Hub struct {
	// Registered clients, grouped by matchID
	matches map[string]map[*Client]bool

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Inbound messages from the clients.
	broadcast chan Message

	mu sync.RWMutex
}

type Message struct {
	MatchID string
	Payload []byte
}

func NewHub() *Hub {
	return &Hub{
		matches:    make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.matches[client.MatchID]; !ok {
				h.matches[client.MatchID] = make(map[*Client]bool)
			}
			h.matches[client.MatchID][client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.matches[client.MatchID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.matches, client.MatchID)
					}
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.matches[message.MatchID]; ok {
				for client := range clients {
					select {
					case client.Send <- message.Payload:
					default:
						close(client.Send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, gm *GameManager, w http.ResponseWriter, r *http.Request, matchID, playerID string) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Create or Get Game
	game := gm.CreateGame(matchID)
	game.AddPlayer(playerID)

	client := &Client{
		Hub:      hub,
		Game:     game,
		MatchID:  matchID,
		PlayerID: playerID,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}
	client.Hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		// Forward input to Game
		var input PlayerInput
		if err := json.Unmarshal(message, &input); err != nil {
			log.Printf("error unmarshalling input: %v", err)
			continue
		}
		// Force PlayerID to match the connection's player ID to prevents spoofing
		input.PlayerID = c.PlayerID

		c.Game.InputChan <- input
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
