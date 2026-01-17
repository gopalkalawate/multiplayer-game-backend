package databases

import (
	"context"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/models"
)

type Database interface {
	CreatePlayer(ctx context.Context, player models.Player) error
	CreateMatch(ctx context.Context, match models.Match) error
	GetMatch(ctx context.Context, playerID string) (models.Match, error)
	ClearTables(ctx context.Context) error
}
