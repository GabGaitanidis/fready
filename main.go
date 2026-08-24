package main

import (
	"log"
	"net/http"

	"fready/internal/config"
	"fready/internal/database"
	"fready/internal/user"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("warning: no .env file loaded:", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}

	conn, err := database.Connect(cfg.DB.DSN())
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer conn.Close()

	log.Println("DB connected successfully with migrations executed")

	userRepo := user.NewRepository(conn)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", userHandler.GetUser)
	mux.HandleFunc("POST /users", userHandler.RegisterUser)
	mux.HandleFunc("GET /users", userHandler.GetUsers)

	log.Printf("Listening on port %s...", cfg.App.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.App.Port, mux))
}