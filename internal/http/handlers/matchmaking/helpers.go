package matchmaking

import (
	"fmt"
	"math"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/models"
)

func CanMatch(p1, p2 models.Player) bool {
	if p1.Region != p2.Region {
		fmt.Println("Players are from different regions")
		return false
	}

	mmrGap := math.Abs(float64(p1.MMR - p2.MMR))
	if mmrGap > 100 {
		fmt.Println("Players have a MMR gap of ", mmrGap)
		return false
	}

	if p1.Ping > p2.Ping {
		fmt.Println("Player 1 has a higher ping than player 2")
		return false
	}

	fmt.Printf("[Match ✅] %s vs %s | MMR Gap: %0.f | Region: %s\n", p1.ID, p2.ID, mmrGap, p1.Region)
	return true
}

func GetTier(mmr int) string {
	if mmr < 500 {
		return "newbie"
	}
	if mmr < 700 {
		return "specialist"
	}
	if mmr < 900 {
		return "expert"
	}
	return "candidate_master"
}

func GetQueueName(region, tier string) string {
	return fmt.Sprintf("queue:%s:%s", region, tier)
}
