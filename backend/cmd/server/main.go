package main

import (
	"context"
	"cyberark-shorten-url/internal/model"
	"cyberark-shorten-url/internal/router"
	"cyberark-shorten-url/internal/storage"
	"cyberark-shorten-url/pkg/logger"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rs/cors"
	"gopkg.in/yaml.v3"
)

func main() {
	configFile := getEnv("CONFIG_FILE", "configs/config.yaml")
	config, loadErr := loadConfig(configFile)
	if loadErr != nil {
		logger.Fatal(fmt.Sprintf("Error loading config: %v", loadErr))
	}

	dbClient := storage.NewDbClient(&config.DbConfig)
	defer dbClient.Close()
	apiRouter := router.NewApiRouter(dbClient, config.SvcConfig.BaseURL, config.SvcConfig.PermanentRedirect, config.SvcConfig.UrlLengthLimit)

	mux := http.NewServeMux()
	apiRouter.RegisterRoutes(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	srv := &http.Server{
		Addr:              config.SvcConfig.Addr,
		Handler:           c.Handler(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// start server on listening
	go func() {
		logger.Info("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server error", "err", err)
		}
	}()

	// monitor signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Warn("shutting down...")

	// gracefully shutdown, wait for max 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	} else {
		logger.Info("server exited successfully")
	}
}

func getEnv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

func loadConfig(configFile string) (*model.Configuration, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg model.Configuration
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&cfg)
	return &cfg, err
}
