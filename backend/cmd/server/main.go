package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/JustCallMeMin/repoCompass/backend/internal/api"
	ghintegration "github.com/JustCallMeMin/repoCompass/backend/internal/integration/github"
	pgstore "github.com/JustCallMeMin/repoCompass/backend/internal/storage/postgres"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	db, err := pgstore.Open(databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	store := pgstore.New(db)
	runner, err := api.NewPersistedLocalRunner(store, logger)
	if err != nil {
		return fmt.Errorf("initialize scan runner: %w", err)
	}

	server := api.NewServer(runner, store, ghintegration.PublicCloner{}, store, logger)
	server.SetGitHubWebhookSecret(os.Getenv("GITHUB_WEBHOOK_SECRET"))
	server.SetDevHeaderAuth(os.Getenv("DEV_HEADER_AUTH") == "true")
	httpServer := &http.Server{
		Addr:              ":" + port(),
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	logger.Info("api server starting", "addr", httpServer.Addr)
	return httpServer.ListenAndServe()
}

func port() string {
	value := os.Getenv("PORT")
	if value == "" {
		return "8080"
	}
	return value
}
