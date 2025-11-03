package main

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/navyn13/redis-go/redis"
)

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func main() {
	loadEnv()
	server := redis.NewServer(redis.Config{
		Username: os.Getenv("USERNAME"),
		Password: os.Getenv("PASSWORD"),
	})

	go func() {
		log.Fatal(server.Start())
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server gracefully...")
	server.Shutdown()
	slog.Info("✅ Server stopped.")
}
