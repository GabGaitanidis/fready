package main

import (
	"fmt"
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

    dbUser, dbPassword, err := config.LoadDbEnvVars()
    if err != nil {
        log.Fatal(err)
    }

    connStr := fmt.Sprintf("postgres://%s:%s@localhost:5432/fready?sslmode=disable", dbUser, dbPassword)
    conn, err := database.Connect(connStr)
    if err != nil {
        log.Println(err)
        return
    } else {
		log.Println("DB connected with migrations being run")
	}

    repo := user.NewRepository(conn)
    service := user.NewService(repo)
    handler := user.NewHandler(service)

    mux := http.NewServeMux()
    mux.HandleFunc("GET /users/{id}", handler.GetUser)
    mux.HandleFunc("POST /users", handler.RegisterUser)
	mux.HandleFunc("GET /users", handler.GetUsers)
    log.Println("Listening on 8080")
    log.Fatal(http.ListenAndServe(":8080", mux))
}