package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/go-faker/faker/v4"
	_ "github.com/lib/pq"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/product"
	"example.com/go-yippi/internal/infrastructure/config"
)

const (
	batchSize  = 1000 // Insert 1000 records per batch
	totalCount = 100000000 // 100 Million records
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	client, err := ent.Open(cfg.Database.Driver, cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed opening connection to database: %v", err)
	}
	defer client.Close()

	// Run auto migration
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	ctx := context.Background()

	// Check existing product count to continue from where we left off
	existingCount, err := client.Product.Query().Count(ctx)
	if err != nil {
		log.Fatalf("failed counting existing products: %v", err)
	}

	if existingCount > 0 {
		log.Printf("Found %d existing products in database. Will continue from SKU-%08d", existingCount, existingCount+1)
	}

	startOffset := existingCount
	actualInsertCount := totalCount

	log.Printf("Starting to seed %d products (%.1fM) from offset %d...", actualInsertCount, float64(actualInsertCount)/1000000, startOffset)
	startTime := time.Now()

	totalBatches := (totalCount + batchSize - 1) / batchSize
	progressInterval := totalBatches / 100 // Report every 1%
	if progressInterval < 1 {
		progressInterval = 1
	}

	// Seed in batches for better performance
	for i := 0; i < actualInsertCount; i += batchSize {
		currentBatch := batchSize
		if i+batchSize > actualInsertCount {
			currentBatch = actualInsertCount - i
		}

		bulk := make([]*ent.ProductCreate, currentBatch)
		for j := 0; j < currentBatch; j++ {
			productNum := startOffset + i + j + 1
			bulk[j] = createProductBuilder(client, productNum)
		}

		// Execute batch insert
		_, err := client.Product.CreateBulk(bulk...).Save(ctx)
		if err != nil {
			log.Fatalf("failed creating products batch %d: %v", i/batchSize+1, err)
		}

		batchNum := i/batchSize + 1

		// Log progress every 1% or for first/last batches
		if batchNum%progressInterval == 0 || batchNum == 1 || batchNum == totalBatches {
			elapsed := time.Since(startTime)
			recordsCompleted := i + currentBatch
			totalRecordsInDB := startOffset + recordsCompleted
			percentComplete := float64(recordsCompleted) / float64(actualInsertCount) * 100
			rate := float64(recordsCompleted) / elapsed.Seconds()

			// Calculate ETA
			remainingRecords := actualInsertCount - recordsCompleted
			etaSeconds := float64(remainingRecords) / rate
			eta := time.Duration(etaSeconds * float64(time.Second))

			log.Printf("Progress: %.1f%% (%d/%d batches) | Inserted: %d | Total in DB: %d | Rate: %.0f/sec | Elapsed: %v | ETA: %v",
				percentComplete,
				batchNum,
				totalBatches,
				recordsCompleted,
				totalRecordsInDB,
				rate,
				elapsed.Round(time.Second),
				eta.Round(time.Second),
			)
		}
	}

	totalDuration := time.Since(startTime)
	finalCount, _ := client.Product.Query().Count(ctx)
	log.Printf("Successfully seeded %d new products in %v", actualInsertCount, totalDuration)
	log.Printf("Average: %.2f products/second", float64(actualInsertCount)/totalDuration.Seconds())
	log.Printf("Total products in database: %d", finalCount)
}

func createProductBuilder(client *ent.Client, num int) *ent.ProductCreate {
	// Generate unique SKU and slug
	sku := fmt.Sprintf("SKU-%08d", num)

	categories := []string{
		"laptop", "phone", "tablet", "monitor", "keyboard",
		"mouse", "headset", "speaker", "camera", "printer",
		"router", "charger", "cable", "adapter", "stand",
		"case", "screen", "drive", "memory", "processor",
	}
	slug := fmt.Sprintf("%s-%d", categories[rand.IntN(len(categories))], num)

	// Generate product name
	brands := []string{
		"TechPro", "SmartDevice", "ProGear", "EliteMax", "PrimeTech",
		"UltraCore", "MegaByte", "PowerEdge", "SwiftTech", "NexGen",
	}
	categoryNames := []string{
		"Laptop", "Smartphone", "Tablet", "Monitor", "Keyboard",
		"Mouse", "Headset", "Speaker", "Camera", "Printer",
	}
	brand := brands[rand.IntN(len(brands))]
	category := categoryNames[rand.IntN(len(categoryNames))]
	name := fmt.Sprintf("%s %s %s", brand, category, randomString(3))

	// Generate realistic price
	price := float64(rand.IntN(4950)+50) + float64(rand.IntN(100))/100.0

	// Generate description
	description := faker.Sentence()

	// Generate shipping dimensions (weight in grams, dimensions in cm)
	weight := rand.IntN(4900) + 100
	length := rand.IntN(40) + 10
	width := rand.IntN(30) + 10
	height := rand.IntN(25) + 5

	// Random status with weighted probability (more published items)
	statusRand := rand.Float64()
	var status product.Status
	switch {
	case statusRand < 0.7: // 70% published
		status = product.StatusPublished
	case statusRand < 0.9: // 20% draft
		status = product.StatusDraft
	default: // 10% archived
		status = product.StatusArchived
	}

	// Random created_at in the past year
	createdAt := time.Now().Add(-time.Duration(rand.IntN(365)) * 24 * time.Hour)
	updatedAt := createdAt.Add(time.Duration(rand.IntN(100)) * 24 * time.Hour)

	return client.Product.Create().
		SetSku(sku).
		SetSlug(slug).
		SetName(name).
		SetPrice(price).
		SetDescription(description).
		SetWeight(weight).
		SetLength(length).
		SetWidth(width).
		SetHeight(height).
		SetStatus(status).
		SetCreatedAt(createdAt).
		SetUpdatedAt(updatedAt)
}

func randomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.IntN(len(charset))]
	}
	return string(result)
}
