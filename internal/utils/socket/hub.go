package socket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected player
type Client struct {
	Hub     *Hub
	MatchID string
	Conn    *websocket.Conn
	Send    chan []byte
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
	/*
	How this works :
	The select statement lets a goroutine wait on multiple communication operations.
	A select blocks until one of its cases can run, then it executes that case. It chooses one at random if multiple are ready.
	
	view it on https://go.dev/tour/concurrency/5
	Select Statement:
		The key is the select statement. Unlike a regular infinite loop that would constantly consume CPU, select blocks until one of its cases can proceed. The goroutine will sit idle, consuming no CPU cycles.
	*/
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
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, matchID string) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{Hub: hub, MatchID: matchID, Conn: conn, Send: make(chan []byte, 256)}
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
		c.Hub.broadcast <- Message{MatchID: c.MatchID, Payload: message}
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
