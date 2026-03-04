package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	api "hw2/.build/openapi"
	"hw2/src/handler"
	"hw2/src/repository"
	"hw2/src/service"
	"hw2/src/middleware"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://myuser:mypassword@localhost:5432/mydb"
	}

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	repo := repository.New(pool)
	svc := service.New(repo)
	h := handler.New(svc)

	srv, err := api.NewServer(h, api.WithErrorHandler(h.ErrorHandler()))
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	fmt.Printf("starting server on %s\n", addr)
	if err := http.ListenAndServe(addr, middleware.Logging(srv)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
