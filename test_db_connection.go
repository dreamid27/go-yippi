package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	dsn := "host=localhost port=5432 user=admin dbname=go-test password=adminadmin sslmode=disable connect_timeout=5"

	fmt.Println("Attempting to connect to database...")
	fmt.Println("DSN:", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open connection: %v", err)
	}
	defer db.Close()

	fmt.Println("Connection opened successfully!")

	// Test ping with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("Testing database ping...")
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	fmt.Println("Ping successful!")

	// Test a simple query
	fmt.Println("Testing simple query...")
	var version string
	if err := db.QueryRowContext(ctx, "SELECT version()").Scan(&version); err != nil {
		log.Fatalf("Failed to query database: %v", err)
	}

	fmt.Println("Query successful!")
	fmt.Println("PostgreSQL version:", version[:50])
	fmt.Println("✓ Database connection is working correctly!")
}
