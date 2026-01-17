package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/databases"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/models"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

// StartMatchmaker listens for new players and attempts to create matches.
// This should be run as a background goroutine.
func StartMatchmaker(db databases.Database) {
	redisClient := utils.GetClient()
	if redisClient == nil {
		log.Println("Redis client is nil, matchmaker cannot start")
		return
	}

	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, "matchmaking_channel")
	defer pubsub.Close()

	ch := pubsub.Channel()

	fmt.Println("Matchmaker started, waiting for players...")

	for msg := range ch {
		fmt.Printf("Received message: %s\n", msg.Payload)
		processQueue(ctx, db)
	}
}

func processQueue(ctx context.Context, db databases.Database) {
	redisClient := utils.GetClient()

	tiers := []string{"newbie", "specialist", "expert", "candidate_master"}
	regions := []string{"US", "EU", "ASIA"} // In production, these should be dynamic or config-based

	for _, region := range regions {
		for _, tier := range tiers {
			queueName := GetQueueName(region, tier)
			processSpecificQueue(ctx, db, redisClient, queueName)
		}
	}
}

func processSpecificQueue(ctx context.Context, db databases.Database, redisClient *redis.Client, queueName string) {
	// Fetch matching players
	// In production, limit this range (e.g. 0-999) and process in batches
	vals, err := redisClient.ZRange(ctx, queueName, 0, -1).Result()
	if err != nil {
		log.Printf("Error reading queue %s: %v\n", queueName, err)
		return
	}

	if len(vals) < 2 {
		return // Not enough players to match in this queue
	}

	// Parse players
	players := make([]models.Player, 0, len(vals))
	for _, v := range vals {
		var p models.Player
		if err := json.Unmarshal([]byte(v), &p); err == nil {
			players = append(players, p)
		}
	}

	// Match logic - O(N) linear scan as players are sorted by MMR
	// Since we are in a specific queue, all players are same region and same tier (roughly).
	// We just need to check if they are "compatible" (Ping, fine-grained MMR gap).

	i := 0
	for i < len(players)-1 {
		p1 := players[i]
		p2 := players[i+1]

		if CanMatch(p1, p2) {
			// Match found!
			createMatch(ctx, db, p1, p2)

			// Remove from Redis
			v1, _ := json.Marshal(p1)
			v2, _ := json.Marshal(p2)

			redisClient.ZRem(ctx, queueName, v1)
			redisClient.ZRem(ctx, queueName, v2)

			// Skip next player
			i += 2
		} else {
			// Try next pair
			i++
		}
	}
}

func createMatch(ctx context.Context, db databases.Database, p1, p2 models.Player) {
	match := models.Match{
		ID:      fmt.Sprintf("%s-%s-%d", p1.ID, p2.ID, time.Now().Unix()),
		Players: []string{p1.ID, p2.ID},
		Region:  p1.Region,
	}

	if err := db.CreateMatch(ctx, match); err != nil {
		log.Printf("Failed to create match in DB: %v\n", err)
	}

	// Here you would typically save the match to a database or Redis
	// and notify the players (e.g., via another Pub/Sub channel or Websocket).
	fmt.Printf("Displaying Match Created: %s between %s and %s\n", match.ID, p1.ID, p2.ID)
}
