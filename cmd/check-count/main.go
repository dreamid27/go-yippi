package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/infrastructure/config"
)

func main() {
	cfg := config.Load()
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed opening connection to database: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	count, err := client.Product.Query().Count(ctx)
	if err != nil {
		log.Fatalf("failed counting products: %v", err)
	}

	fmt.Printf("Total products in database: %d\n", count)

	// Get the last product to show details
	if count > 0 {
		lastProduct, err := client.Product.Query().
			Order(ent.Desc("created_at")).
			First(ctx)
		if err != nil {
			log.Fatalf("failed getting last product: %v", err)
		}
		fmt.Printf("Last Product: %s (Slug: %s)\n", lastProduct.Name, lastProduct.Slug)
	}
}
