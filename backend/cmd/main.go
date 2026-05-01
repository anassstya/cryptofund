package main

import (
	"context"
	"cryptofund/db"
	"cryptofund/internal/auth"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")

	if databaseURL == "" || jwtSecret == "" {
		log.Fatal("DATABASE_URL and JWT_SECRET must be set in environment")
	}

	if err := db.MigrateDB(databaseURL, "/app/migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, jwtSecret)
	authHdl := auth.NewHandler(authSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", authHdl.RegisterHandler)
	//mux.HandleFunc("POST /login", authHdl.LoginHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

//СHI add
