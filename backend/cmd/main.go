package main

import (
	"context"
	"cryptofund/db"
	"cryptofund/internal/auth"
	"cryptofund/internal/exchanges"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	jwtSecret := os.Getenv("JWT_SECRET")
	masterKey := os.Getenv("ENCRYPTION_MASTER_KEY")

	if databaseURL == "" || jwtSecret == "" {
		log.Fatal("DATABASE_URL and JWT_SECRET must be set in environment")
	}

	ctx := context.Background()

	var pool *pgxpool.Pool
	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		pool, err = pgxpool.New(ctx, databaseURL)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				log.Println("Database connected successfully!")
				break
			}
			pool.Close()
		}
		log.Printf("Waiting for database... (attempt %d/%d): %v", i+1, maxRetries, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to database after %d attempts: %v", maxRetries, err)
	}
	defer pool.Close()

	if err := db.MigrateDB(databaseURL, "/app/migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migrations applied successfully")

	authRepo := auth.NewRepository(pool)
	authSvc := auth.NewService(authRepo, jwtSecret)
	authHdl := auth.NewHandler(authSvc)

	exchangeRepo := exchanges.NewRepository(pool)
	exchangeSvc := exchanges.NewService(exchangeRepo, masterKey)
	exchangeHdl := exchanges.NewHandler(exchangeSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", authHdl.RegisterHandler)
	mux.HandleFunc("POST /login", authHdl.LoginHandler)

	mux.HandleFunc("POST /exchange", auth.MiddlewareAuth(exchangeHdl.AddExchangeHandler))

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
