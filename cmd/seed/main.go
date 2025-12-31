package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/google/uuid"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/product"
	"example.com/go-yippi/internal/infrastructure/config"
)

const (
	batchSize  = 1000 // Insert 1000 records per batch
	totalCount = 100000000 // 100 Million records
)

// Realistic tech brands for e-commerce
var techBrands = []string{
	"Apple", "Samsung", "Dell", "HP", "Lenovo", "ASUS", "MSI", "Acer",
	"LG", "Sony", "Microsoft", "Google", "Xiaomi", "OnePlus", "Oppo",
	"Vivo", "Realme", "Huawei", "ZTE", "Motorola", "Nokia", "BlackBerry",
	"Toshiba", "Panasonic", "Sharp", "Fujitsu", "Razer", "Corsair", "Logitech",
	"JBL", "Bose", "Sennheiser", "Audio-Technica", "Shure", "Plantronics",
	"Cisco", "Netgear", "TP-Link", "D-Link", "ASRock", "Gigabyte", "Biostar",
	"Crucial", "Kingston", "SanDisk", "Western Digital", "Seagate",
	"Epson", "Canon", "Brother", "Xerox", "Fujifilm", "Nikon",
}

// Hierarchical category structure for electronics
var categoryStructure = map[string][]string{
	"Computers": {
		"Laptops", "Desktop PCs", "Tablets", "Monitors", "Keyboards", "Mice", "Printers", "Scanners",
	},
	"Mobile Devices": {
		"Smartphones", "Feature Phones", "Smartwatches", "Fitness Trackers", "Mobile Accessories",
	},
	"Audio": {
		"Headphones", "Earbuds", "Speakers", "Soundbars", "Audio Systems", "Microphones",
	},
	"Gaming": {
		"Gaming Laptops", "Gaming Desktops", "Gaming Consoles", "Gaming Chairs", "Gaming Desks", "Gaming Accessories",
	},
	"Networking": {
		"Routers", "Switches", "Access Points", "Network Cables", "Modems", "Range Extenders",
	},
	"Storage": {
		"External Hard Drives", "Internal Hard Drives", "SSD", "USB Flash Drives", "Memory Cards", "NAS",
	},
	"Components": {
		"Processors", "Graphics Cards", "Motherboards", "RAM", "Power Supplies", "Cooling Systems", "Cases",
	},
	"Cameras": {
		"DSLR Cameras", "Mirrorless Cameras", "Action Cameras", "Security Cameras", "Webcams", "Camera Accessories",
	},
}

// Product name templates for each category
var productTemplates = map[string][]string{
	"Laptops":        {"%s %s Pro", "%s %s Air", "%s %s Ultra", "%s %s Gaming", "%s %s Business"},
	"Desktop PCs":    {"%s %s Tower", "%s %s Mini", "%s %s Workstation", "%s %s Gaming", "%s %s All-in-One"},
	"Smartphones":    {"%s %s Pro", "%s %s Plus", "%s %s Lite", "%s %s Ultra", "%s %s SE"},
	"Tablets":        {"%s %s Pro", "%s %s Air", "%s %s Mini", "%s %s Ultra", "%s %s Lite"},
	"Monitors":       {"%s %s Display", "%s %s Pro", "%s %s Ultra", "%s %s Gaming", "%s %s 4K"},
	"Headphones":     {"%s %s Pro", "%s %s Max", "%s %s Sport", "%s %s Studio", "%s %s Wireless"},
	"Speakers":       {"%s %s Smart", "%s %s Portable", "%s %s Studio", "%s %s Party", "%s %s Home"},
	"Keyboards":      {"%s %s Mechanical", "%s %s Wireless", "%s %s Gaming", "%s %s Ergonomic", "%s %s Compact"},
	"Mice":           {"%s %s Gaming", "%s %s Wireless", "%s %s Ergonomic", "%s %s Trackball", "%s %s Vertical"},
	"Printers":       {"%s %s Laser", "%s %s Inkjet", "%s %s All-in-One", "%s %s Photo", "%s %s Office"},
	"Smartwatches":   {"%s %s Pro", "%s %s Sport", "%s %s Classic", "%s %s Active", "%s %s Ultra"},
	"Routers":        {"%s %s Pro", "%s %s Gaming", "%s %s Mesh", "%s %s Business", "%s %s Home"},
	"Processors":     {"%s %s Series", "%s %s X", "%s %s K", "%s %s F", "%s %s G"},
	"Graphics Cards": {"%s %s RTX", "%s %s Radeon", "%s %s Gaming", "%s %s Pro", "%s %s Workstation"},
	"RAM":            {"%s %s DDR4", "%s %s DDR5", "%s %s Gaming", "%s %s RGB", "%s %s Server"},
}

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

	// Create brands first
	log.Printf("Creating brands...")
	brandMap, err := createBrands(ctx, client)
	if err != nil {
		log.Fatalf("failed creating brands: %v", err)
	}
	log.Printf("Created %d brands", len(brandMap))

	// Create categories
	log.Printf("Creating categories...")
	categoryMap, err := createCategories(ctx, client)
	if err != nil {
		log.Fatalf("failed creating categories: %v", err)
	}
	log.Printf("Created %d categories", len(categoryMap))

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
			bulk[j] = createProductBuilder(client, productNum, brandMap, categoryMap)
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

// createBrands creates all tech brands and returns a map of brand name to UUID
func createBrands(ctx context.Context, client *ent.Client) (map[string]uuid.UUID, error) {
	brandMap := make(map[string]uuid.UUID)

	// First, try to find existing brands
	existingBrands, err := client.Brand.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing brands: %w", err)
	}

	// Map existing brands
	for _, brand := range existingBrands {
		brandMap[brand.Name] = brand.ID
	}

	// Create missing brands
	for _, brandName := range techBrands {
		if _, exists := brandMap[brandName]; !exists {
			brand, err := client.Brand.Create().
				SetName(brandName).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create brand %s: %w", brandName, err)
			}
			brandMap[brandName] = brand.ID
		}
	}

	return brandMap, nil
}

// createCategories creates hierarchical categories and returns a map of category name to UUID
func createCategories(ctx context.Context, client *ent.Client) (map[string]uuid.UUID, error) {
	categoryMap := make(map[string]uuid.UUID)

	// First, try to find existing categories
	existingCategories, err := client.Category.Query().All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing categories: %w", err)
	}

	// Map existing categories
	for _, category := range existingCategories {
		categoryMap[category.Name] = category.ID
	}

	// Create missing parent categories
	for parentName := range categoryStructure {
		if _, exists := categoryMap[parentName]; !exists {
			parent, err := client.Category.Create().
				SetName(parentName).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to create parent category %s: %w", parentName, err)
			}
			categoryMap[parentName] = parent.ID
		}
	}

	// Create missing child categories and link to parents
	for parentName, children := range categoryStructure {
		parentID := categoryMap[parentName]

		for _, childName := range children {
			if _, exists := categoryMap[childName]; !exists {
				child, err := client.Category.Create().
					SetName(childName).
					SetParentID(parentID).
					Save(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to create child category %s: %w", childName, err)
				}
				categoryMap[childName] = child.ID
			}
		}
	}

	return categoryMap, nil
}

// getLeafCategories returns all leaf categories (categories without children) for product assignment
func getLeafCategories() []string {
	var leafCategories []string
	for _, children := range categoryStructure {
		leafCategories = append(leafCategories, children...)
	}
	return leafCategories
}

// generateProductDescription generates a realistic description based on category
func generateProductDescription(category, brand string) string {
	descriptions := map[string]string{
		"Laptops":        fmt.Sprintf("High-performance laptop from %s with latest generation processor, vibrant display, and long battery life. Perfect for work and entertainment.", brand),
		"Smartphones":    fmt.Sprintf("Premium smartphone from %s featuring advanced camera system, stunning display, and powerful performance. Capture every moment in brilliant detail.", brand),
		"Tablets":        fmt.Sprintf("Versatile tablet from %s with responsive display, powerful processor, and all-day battery. Ideal for work, creativity, and entertainment.", brand),
		"Monitors":       fmt.Sprintf("Professional display from %s with stunning resolution, accurate colors, and ergonomic design. Perfect for productivity and creative work.", brand),
		"Headphones":     fmt.Sprintf("Premium audio experience from %s with active noise cancellation, superior sound quality, and comfortable all-day wear.", brand),
		"Speakers":       fmt.Sprintf("Powerful sound system from %s with immersive audio, deep bass, and wireless connectivity. Transform any space into a concert hall.", brand),
		"Keyboards":      fmt.Sprintf("Precision keyboard from %s with responsive keys, customizable backlighting, and ergonomic design for enhanced typing experience.", brand),
		"Mice":           fmt.Sprintf("Ergonomic mouse from %s with precise tracking, customizable buttons, and comfortable grip for extended use.", brand),
		"Printers":       fmt.Sprintf("Reliable printer from %s with fast printing, high-quality output, and wireless connectivity for home and office use.", brand),
		"Smartwatches":   fmt.Sprintf("Advanced smartwatch from %s with health monitoring, fitness tracking, and seamless connectivity. Stay connected and healthy.", brand),
		"Routers":        fmt.Sprintf("High-performance router from %s with fast Wi-Fi, wide coverage, and advanced security for all your connected devices.", brand),
		"Processors":     fmt.Sprintf("Powerful processor from %s with advanced performance, energy efficiency, and support for the latest computing technologies.", brand),
		"Graphics Cards": fmt.Sprintf("High-performance graphics card from %s with advanced rendering, ray tracing, and AI capabilities for gaming and content creation.", brand),
		"RAM":            fmt.Sprintf("High-speed memory module from %s with reliable performance, optimized latency, and perfect for multitasking and demanding applications.", brand),
	}

	if desc, exists := descriptions[category]; exists {
		return desc
	}

	// Default description
	return fmt.Sprintf("Quality product from %s designed for performance and reliability. Built with premium materials and advanced technology.", brand)
}

// generateRealisticPrice generates realistic pricing based on category
func generateRealisticPrice(category string) float64 {
	priceRanges := map[string]struct {
		min, max float64
	}{
		"Laptops":         {500, 2500},
		"Desktop PCs":     {400, 3000},
		"Smartphones":     {200, 1500},
		"Tablets":         {150, 1200},
		"Monitors":        {150, 2000},
		"Headphones":      {30, 500},
		"Speakers":        {25, 800},
		"Keyboards":       {20, 300},
		"Mice":            {15, 150},
		"Printers":        {50, 600},
		"Smartwatches":    {100, 800},
		"Routers":         {30, 400},
		"Processors":      {100, 1000},
		"Graphics Cards":  {150, 2000},
		"RAM":             {25, 300},
		"Gaming Laptops":  {800, 3500},
		"Gaming Desktops": {600, 4000},
		"Gaming Consoles": {250, 600},
		"SSD":             {30, 500},
		"External Hard Drives": {40, 300},
		"DSLR Cameras":    {400, 3000},
		"Security Cameras": {50, 400},
	}

	if prange, exists := priceRanges[category]; exists {
		return float64(rand.IntN(int(prange.max-prange.min)*100)+int(prange.min)*100) / 100.0
	}

	// Default price range
	return float64(rand.IntN(450)+50) + float64(rand.IntN(100))/100.0
}

// generateRealisticDimensions generates realistic dimensions based on category
func generateRealisticDimensions(category string) (weight, length, width, height int) {
	switch {
	case strings.Contains(category, "Laptop"):
		weight = rand.IntN(1500) + 1000      // 1-2.5 kg
		length = rand.IntN(10) + 30          // 30-40 cm
		width = rand.IntN(8) + 20             // 20-28 cm
		height = rand.IntN(3) + 1             // 1-4 cm

	case strings.Contains(category, "Phone"):
		weight = rand.IntN(100) + 150         // 150-250 g
		length = rand.IntN(20) + 130         // 130-150 cm
		width = rand.IntN(15) + 60           // 60-75 cm
		height = rand.IntN(10) + 6           // 6-16 cm

	case strings.Contains(category, "Tablet"):
		weight = rand.IntN(300) + 300         // 300-600 g
		length = rand.IntN(50) + 180         // 180-230 cm
		width = rand.IntN(30) + 120          // 120-150 cm
		height = rand.IntN(8) + 4            // 4-12 cm

	case strings.Contains(category, "Monitor"):
		weight = rand.IntN(3000) + 2000       // 2-5 kg
		length = rand.IntN(30) + 50           // 50-80 cm
		width = rand.IntN(20) + 30            // 30-50 cm
		height = rand.IntN(40) + 10           // 10-50 cm

	case strings.Contains(category, "Keyboard"):
		weight = rand.IntN(500) + 500         // 500-1000 g
		length = rand.IntN(20) + 35           // 35-55 cm
		width = rand.IntN(10) + 10            // 10-20 cm
		height = rand.IntN(5) + 2             // 2-7 cm

	case strings.Contains(category, "Mouse"):
		weight = rand.IntN(100) + 50          // 50-150 g
		length = rand.IntN(5) + 10            // 10-15 cm
		width = rand.IntN(5) + 5             // 5-10 cm
		height = rand.IntN(5) + 2             // 2-7 cm

	case strings.Contains(category, "Headphone"):
		weight = rand.IntN(200) + 200         // 200-400 g
		length = rand.IntN(10) + 15           // 15-25 cm
		width = rand.IntN(15) + 10           // 10-25 cm
		height = rand.IntN(15) + 10           // 10-25 cm

	case strings.Contains(category, "Speaker"):
		weight = rand.IntN(2000) + 500        // 500-2500 g
		length = rand.IntN(20) + 20           // 20-40 cm
		width = rand.IntN(20) + 15           // 15-35 cm
		height = rand.IntN(30) + 20           // 20-50 cm

	default:
		// Default dimensions
		weight = rand.IntN(2000) + 100         // 100-2100 g
		length = rand.IntN(30) + 10           // 10-40 cm
		width = rand.IntN(25) + 5             // 5-30 cm
		height = rand.IntN(20) + 5             // 5-25 cm
	}

	return weight, length, width, height
}

func createProductBuilder(client *ent.Client, num int, brandMap map[string]uuid.UUID, categoryMap map[string]uuid.UUID) *ent.ProductCreate {
	// Generate unique SKU and slug
	sku := fmt.Sprintf("SKU-%08d", num)

	// Select a random leaf category for product assignment
	leafCategories := getLeafCategories()
	selectedCategory := leafCategories[rand.IntN(len(leafCategories))]
	categoryID := categoryMap[selectedCategory]

	// Select a random brand
	selectedBrand := techBrands[rand.IntN(len(techBrands))]
	brandID := brandMap[selectedBrand]

	// Generate product name using templates or fallback
	productName := generateProductName(selectedBrand, selectedCategory)
	slug := fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(productName, " ", "-")), num)

	// Generate realistic price based on category
	price := generateRealisticPrice(selectedCategory)

	// Generate realistic description
	description := generateProductDescription(selectedCategory, selectedBrand)

	// Generate realistic dimensions
	weight, length, width, height := generateRealisticDimensions(selectedCategory)

	// Generate sample image URLs (placeholder)
	imageUrls := []string{
		fmt.Sprintf("https://example.com/images/products/%d-1.jpg", num),
		fmt.Sprintf("https://example.com/images/products/%d-2.jpg", num),
		fmt.Sprintf("https://example.com/images/products/%d-3.jpg", num),
	}

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
		SetName(productName).
		SetPrice(price).
		SetDescription(description).
		SetWeight(weight).
		SetLength(length).
		SetWidth(width).
		SetHeight(height).
		SetImageUrls(imageUrls).
		SetStatus(status).
		SetCategoryID(categoryID).
		SetBrandID(brandID).
		SetCreatedAt(createdAt).
		SetUpdatedAt(updatedAt)
}

func generateProductName(brandName, category string) string {
	// Try to use templates first
	if templates, exists := productTemplates[category]; exists {
		template := templates[rand.IntN(len(templates))]

		// Generate model number/identifier
		modelIdentifiers := []string{"X1", "Pro Max", "Ultra", "Elite", "Plus", "Premium", "Advanced", "Supreme"}
		modelID := modelIdentifiers[rand.IntN(len(modelIdentifiers))]

		// Add random numbers for variation
		modelNumber := rand.IntN(9000) + 1000

		// Combine template parts
		if strings.Contains(template, "%s %s") {
			return fmt.Sprintf(template, brandName, modelID)
		} else {
			return fmt.Sprintf("%s %s %s %d", brandName, category, modelID, modelNumber)
		}
	}

	// Fallback for categories without templates
	return fmt.Sprintf("%s %s %d", brandName, category, rand.IntN(9000)+1000)
}
