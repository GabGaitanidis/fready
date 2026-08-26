package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"fready/internal/auth"
	"fready/internal/auth/session"
	"fready/internal/config"
	"fready/internal/database"
	"fready/internal/group"
	"fready/internal/middlewares"
	"fready/internal/user"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With(
		"service", "fready",
	)
	slog.SetDefault(logger)
	if err := godotenv.Load(); err != nil {
		log.Println("warning: no .env file loaded:", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	
	conn, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	slog.Info("database connected")
	defer conn.Close()

	log.Println("DB connected successfully with migrations executed")
	mux := http.NewServeMux()

    userService := user.NewRouter().RegisterRoutes(mux, conn)

    sessionRepo := session.NewSessionRepository(conn)
    sessionService := session.NewService(sessionRepo)
	groupService := group.NewRouter().RegisterRoutes(mux, conn, sessionService)
	_ = groupService
    auth.NewRouter().RegisterRoutes(mux, userService, sessionService)


	slog.Info("server starting", "port", "8080")
	log.Fatal(http.ListenAndServe(":8080", middlewares.LoggingMiddleware(mux)))
}