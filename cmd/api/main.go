package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"find/internal/httpapi"
	"find/internal/store"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://find:find@127.0.0.1:5433/find?sslmode=disable"
	}
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		dir = "migrations"
	}

	ctx := context.Background()
	var db *store.Store
	var err error
	for attempt := 1; attempt <= 30; attempt++ {
		db, err = store.Open(ctx, dsn)
		if err == nil {
			break
		}
		log.Printf("postgres: %v", err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Migrate(ctx, dir); err != nil {
		log.Fatal(err)
	}
	if err := db.Seed(ctx); err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:              "127.0.0.1:8080",
		Handler:           httpapi.New(db).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("api http://%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
