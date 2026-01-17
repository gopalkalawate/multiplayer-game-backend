package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/config"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*SQLite, error) {
	db, err := sql.Open("sqlite3", cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	return &SQLite{Db: db}, nil
}

func (s *SQLite) CreatePlayer(ctx context.Context, player models.Player) error {
	query := `INSERT INTO players (id, mmr, ping, region, tier, created_at) VALUES (?, ?, ?, ?, ?, ?)`

	tier := "candidate_master"
	if player.MMR < 500 {
		tier = "newbie"
	} else if player.MMR < 700 {
		tier = "specialist"
	} else if player.MMR < 900 {
		tier = "expert"
	}

	stmt, err := s.Db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		player.ID,
		player.MMR,
		player.Ping,
		player.Region,
		tier,
		time.Unix(player.JoinedAt, 0),
	)
	return err
}

func (s *SQLite) CreateMatch(ctx context.Context, match models.Match) error {
	/*
		We need this to be atomic. If match is created but players are not inserted, then rollback
		else commit
	*/
	tx, err := s.Db.BeginTx(ctx, nil) // begin transaction
	if err != nil {
		return err
	}
	defer tx.Rollback() // if not committed, rollback

	// Insert Match
	matchQuery := `INSERT INTO matches (id, created_at) VALUES (?, CURRENT_TIMESTAMP)`
	if _, err := tx.ExecContext(ctx, matchQuery, match.ID); err != nil {
		return err
	}

	// Insert Match Players
	playerQuery := `INSERT INTO matches_players (match_id, player_id) VALUES (?, ?)`
	statusQuery := `UPDATE players SET status = 'matched' WHERE id = ?`

	stmt, err := tx.PrepareContext(ctx, playerQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	statusStmt, err := tx.PrepareContext(ctx, statusQuery)
	if err != nil {
		return err
	}
	defer statusStmt.Close()

	for _, playerID := range match.Players {
		if _, err := stmt.ExecContext(ctx, match.ID, playerID); err != nil {
			return err
		}
		if _, err := statusStmt.ExecContext(ctx, playerID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLite) GetMatch(ctx context.Context, playerID string) (models.Match, error) {
	// Check player status
	var status string
	err := s.Db.QueryRowContext(ctx, "SELECT status FROM players WHERE id = ?", playerID).Scan(&status)
	if err != nil {
		return models.Match{}, err
	}

	if status == "waiting" {
		return models.Match{Status: "waiting"}, nil
	}

	query := `SELECT match_id FROM matches_players WHERE player_id = ?`
	row := s.Db.QueryRowContext(ctx, query, playerID)
	var matchID string
	if err := row.Scan(&matchID); err != nil {
		return models.Match{}, err
	}
	return models.Match{ID: matchID, Status: "matched"}, nil
}

func (s *SQLite) ClearTables(ctx context.Context) error {
	tx, err := s.Db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tables := []string{"matches_players", "matches", "players"} // Order matters due to FKs if any
	for _, table := range tables {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return err
		}
	}

	return tx.Commit()
}
