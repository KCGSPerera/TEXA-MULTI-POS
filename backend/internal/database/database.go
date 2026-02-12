package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(databaseURL string) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatal("Unable to parse DATABASE_URL:", err)
	}

	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour

	DB, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}

	log.Println("✅ Connected to Supabase PostgreSQL")
}
