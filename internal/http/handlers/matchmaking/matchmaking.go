package matchmaking

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/databases"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/models"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/utils"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/utils/response"
	"github.com/redis/go-redis/v9"
)

func JoinQueue(db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var player models.Player
		fmt.Println("match request received")
		err := json.NewDecoder(r.Body).Decode(&player)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		fmt.Println("player decoded")

		ctx := r.Context()

		// Persist player to database first
		// if err := db.CreatePlayer(ctx, player); err != nil {
		// 	fmt.Printf("Error creating player in DB: %v\n", err)
		// 	response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("failed to persist player: %w", err)))
		// 	return
		// }

		redisClient := utils.GetClient()
		if redisClient == nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(fmt.Errorf("redis client is nil")))
			return
		}

		fmt.Println("redis client initialized")

		// Marshal player to JSON
		pBytes, err := json.Marshal(player)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		// Enqueue player
		tier := GetTier(player.MMR)
		queueName := GetQueueName(player.Region, tier) // e.g. queue:US:newbie

		fmt.Printf("enqueueing player to %s\n", queueName)

		err = redisClient.ZAdd(ctx, queueName, redis.Z{
			Score:  float64(player.MMR),
			Member: pBytes,
		}).Err()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		fmt.Println("player enqueued to ZSET, publishing new player event")

		// Publish event
		err = redisClient.Publish(ctx, "matchmaking_channel", "new_player").Err() // something changed. Wake up Suscriber
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}
		fmt.Println("new player event published")
		response.WriteJson(w, http.StatusOK, response.SuccessResponse{
			Status: "waiting for match",
			Data:   nil,
		})
	}
}

func GetMatchStatus(db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var player models.Player
		err := json.NewDecoder(r.Body).Decode(&player)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		ctx := r.Context()

		match, err := db.GetMatch(ctx, player.ID)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, response.SuccessResponse{
			Status: "match found",
			Data:   match,
		})
	}
}
