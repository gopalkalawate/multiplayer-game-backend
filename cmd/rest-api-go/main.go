package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gopalkalawate/multiplayer-game-backend/internal/config"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/databases/sqlite"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/http/handlers/matchmaking"
	"github.com/gopalkalawate/multiplayer-game-backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

func main() {
	// load configs
	cfg := config.MustLoad()

	// setup database
	db, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// setup redis
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	utils.SetClient(rdb)

	slog.Info("Storage Initialized", slog.String("env", cfg.Env))

	// start matchmaker worker
	go matchmaking.StartMatchmaker(db)
	// setup router
	router := http.NewServeMux()
	router.HandleFunc("POST /join-queue", matchmaking.JoinQueue(db))
	router.HandleFunc("GET /match-status", matchmaking.GetMatchStatus(db))

	server := &http.Server{
		Addr:    cfg.HTTPServer.Address,
		Handler: router,
	}

	slog.Info("Server starting", slog.String("addr", cfg.HTTPServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	slog.Info("Server Started")

	<-done
	slog.Info("Server Stopped")

	if cfg.Env == "local" || cfg.Env == "dev" {
		slog.Info("Cleaning up database and redis...")
		if err := db.ClearTables(context.Background()); err != nil {
			slog.Error("Failed to clear tables", slog.String("error", err.Error()))
		}

		if err := rdb.FlushAll(context.Background()).Err(); err != nil {
			slog.Error("Failed to flush redis", slog.String("error", err.Error()))
		}
		slog.Info("Cleanup complete")
	}

	if err := server.Shutdown(context.Background()); err != nil {
		slog.Error("Server Shutdown Failed", slog.String("error", err.Error()))
	}
	slog.Info("Server Exited Properly")
}
