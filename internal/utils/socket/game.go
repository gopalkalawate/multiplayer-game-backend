package socket

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

// GameManager manages the state of all active games
type GameManager struct {
	Games map[string]*Game
	Hub   *Hub
	mu    sync.RWMutex
}

func NewGameManager(hub *Hub) *GameManager {
	return &GameManager{
		Games: make(map[string]*Game),
		Hub:   hub,
	}
}

func (gm *GameManager) CreateGame(matchID string) *Game {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.Games[matchID]; exists {
		return gm.Games[matchID]
	}

	game := &Game{
		MatchID: matchID,
		State: GameState{
			MatchID: matchID,
			Players: make(map[string]*CarState),
		},
		Hub:       gm.Hub,
		InputChan: make(chan PlayerInput),
		Ctx:       context.Background(),
	}

	// Start game loop
	go game.Run()

	gm.Games[matchID] = game
	log.Printf("Game created for match %s", matchID)
	return game
}

// Game represents a single running match
type Game struct {
	MatchID   string
	State     GameState
	Hub       *Hub
	InputChan chan PlayerInput
	Ctx       context.Context
	mu        sync.RWMutex
}

type GameState struct {
	MatchID string               `json:"match_id"`
	Tick    int64                `json:"tick"`
	Players map[string]*CarState `json:"players"`
}

type CarState struct {
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Width        float64 `json:"width"`
	Height       float64 `json:"height"`
	Speed        float64 `json:"speed"`
	Acceleration float64 `json:"acceleration"`
	MaxSpeed     float64 `json:"maxSpeed"`
	Friction     float64 `json:"friction"`
	Angle        float64 `json:"angle"`
	Damaged      bool    `json:"damaged"`
}

type PlayerInput struct {
	PlayerID string   `json:"player_id"`
	Action   string   `json:"action"`
	Payload  CarState `json:"payload"`
}

func (g *Game) Run() {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-g.Ctx.Done():
			return
		case <-ticker.C:
			g.mu.Lock()
			g.State.Tick++
			// Here: Update Physics using Speed, Acceleration, Angle etc.
			// loop through players and update positions based on speed/angle if server authoritative.

			// Broadcast state
			stateBytes, _ := json.Marshal(g.State)
			g.mu.Unlock()

			// Send to Hub to broadcast to specific room
			g.Hub.broadcast <- Message{MatchID: g.MatchID, Payload: stateBytes}

		case input := <-g.InputChan:
			g.mu.Lock()
			if player, ok := g.State.Players[input.PlayerID]; ok {
				// Update player state from input payload (Client Authoritative for now)
				player.X = input.Payload.X
				player.Y = input.Payload.Y
				player.Speed = input.Payload.Speed
				player.Angle = input.Payload.Angle
				// We can update other fields if sent
			}
			g.mu.Unlock()
		}
	}
}

// Helper to add player
func (g *Game) AddPlayer(playerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Initialize default car state
	g.State.Players[playerID] = &CarState{
		X:            0,
		Y:            0,
		Width:        20, // Example defaults
		Height:       40,
		Speed:        0,
		Acceleration: 0.1,
		MaxSpeed:     10, // Example
		Friction:     0.05,
		Angle:        0,
		Damaged:      false,
	}
}
