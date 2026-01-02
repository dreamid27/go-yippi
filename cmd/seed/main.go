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
	"golang.org/x/crypto/bcrypt"

	"example.com/go-yippi/internal/adapters/persistence/db/ent"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/product"
	"example.com/go-yippi/internal/adapters/persistence/db/ent/user"
	"example.com/go-yippi/internal/domain/entities"
	"example.com/go-yippi/internal/infrastructure/config"
)

const (
	batchSize  = 100 // Insert 100 records per batch
	totalCount = 100 // 100 products for testing
)

// Fashion brands for e-commerce
var fashionBrands = []string{
	"Zara", "H&M", "Uniqlo", "Gucci", "Prada", "Nike", "Adidas", "Puma",
	"Levi's", "Calvin Klein", "Tommy Hilfiger", "Ralph Lauren", "GAP", "Banana Republic",
	"Mango", "Topshop", "Forever 21", "H&M", "Massimo Dutti", "Pull&Bear",
	"Bershka", "Stradivarius", "Oysho", "Zara Home", "Lefties", "Uterqüe",
	"Lacoste", "Fred Perry", "Superdry", "The North Face", "Columbia", "Patagonia",
	"Timberland", "Dr. Martens", "Converse", "Vans", "New Balance", "Reebok",
	"Fila", "Champion", "Diesel", "Guess", "Armani", "Versace", "Dolce & Gabbana",
	"Burberry", "Balenciaga", "Saint Laurent", "Givenchy", "Valentino", "Fendi",
	"Moncler", "Canada Goose", "Woolrich", "Arc'teryx", "Salomon",
	"IKEA", "Muji", "Daiso", "Miniso", "Cotton On", "Uniqlo",
}

// Hierarchical category structure for fashion
var categoryStructure = map[string][]string{
	"Women": {
		"Dresses", "Tops & Blouses", "Pants & Jeans", "Skirts", "Jackets & Coats", "Knitwear", "Blazers", "Lingerie",
	},
	"Men": {
		"T-Shirts & Polos", "Shirts", "Pants & Jeans", "Jackets & Coats", "Suits & Blazers", "Knitwear", "Shorts", "Underwear",
	},
	"Shoes": {
		"Sneakers", "Boots", "Heels", "Flats", "Sandals", "Loafers", "Athletic Shoes", "Formal Shoes",
	},
	"Accessories": {
		"Bags & Handbags", "Wallets", "Belts", "Scarves", "Hats & Caps", "Jewelry", "Watches", "Sunglasses",
	},
	"Sportswear": {
		"Athletic Wear", "Yoga & Fitness", "Running Gear", "Swimwear", "Sports Accessories", "Outdoor Wear",
	},
	"Kids": {
		"Girls' Clothing", "Boys' Clothing", "Kids' Shoes", "Baby Clothing", "Kids' Accessories",
	},
	"Loungewear": {
		"Pajamas", "Robes", "Sleepwear", "Loungewear Sets", "Slippers",
	},
	"Workwear": {
		"Uniforms", "Scrubs", "Work Pants", "Work Jackets", "Safety Shoes",
	},
}

// Product name templates for each category
var productTemplates = map[string][]string{
	"Dresses":            {"%s %s Dress", "%s %s Maxi Dress", "%s %s Mini Dress", "%s %s Midi Dress", "%s %s Wrap Dress"},
	"Tops & Blouses":     {"%s %s Blouse", "%s %s Top", "%s %s Shirt", "%s %s T-Shirt", "%s %s Polo"},
	"Pants & Jeans":      {"%s %s Jeans", "%s %s Pants", "%s %s Trousers", "%s %s Chinos", "%s %s Leggings"},
	"Skirts":             {"%s %s Skirt", "%s %s Mini Skirt", "%s %s Midi Skirt", "%s %s Pencil Skirt", "%s %s A-Line Skirt"},
	"Jackets & Coats":    {"%s %s Jacket", "%s %s Coat", "%s %s Blazer", "%s %s Parka", "%s %s Trench Coat"},
	"Knitwear":           {"%s %s Sweater", "%s %s Cardigan", "%s %s Pullover", "%s %s Hoodie", "%s %s Jumper"},
	"Sneakers":           {"%s %s Sneaker", "%s %s Trainer", "%s %s Running Shoe", "%s %s Basketball Shoe", "%s %s Skate Shoe"},
	"Boots":              {"%s %s Boots", "%s %s Ankle Boots", "%s %s Knee Boots", "%s %s Chelsea Boots", "%s %s Combat Boots"},
	"Heels":              {"%s %s Heels", "%s %s Stilettos", "%s %s Pumps", "%s %s Platforms", "%s %s Wedges"},
	"Bags & Handbags":    {"%s %s Handbag", "%s %s Tote Bag", "%s %s Crossbody Bag", "%s %s Shoulder Bag", "%s %s Clutch"},
	"Watches":            {"%s %s Watch", "%s %s Chronograph", "%s %s Digital Watch", "%s %s Analog Watch", "%s %s Smart Watch"},
	"Sunglasses":         {"%s %s Sunglasses", "%s %s Aviators", "%s %s Wayfarers", "%s %s Round Frames", "%s %s Cat Eye"},
	"Athletic Wear":      {"%s %s Jersey", "%s %s Shorts", "%s %s Tracksuit", "%s %s Leggings", "%s %s Tank Top"},
	"Swimwear":           {"%s %s Bikini", "%s %s One-Piece", "%s %s Swimsuit", "%s %s Board Shorts", "%s %s Rash Guard"},
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

	// Create admin user first
	log.Printf("Creating admin user...")
	err = createAdminUser(ctx, client)
	if err != nil {
		log.Fatalf("failed creating admin user: %v", err)
	}
	log.Printf("Admin user created successfully")

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

	// Create variants for products
	log.Printf("Starting to create product variants...")
	variantStartTime := time.Now()

	// Get all products that should have variants (70% of products)
	products, err := client.Product.Query().
		WithCategory(). // Fetch category relationship
		All(ctx)
	if err != nil {
		log.Fatalf("failed querying products for variant creation: %v", err)
	}

	// All products should have variants in fashion e-commerce
	var productsNeedingVariants []*ent.Product
	productsNeedingVariants = append(productsNeedingVariants, products...)

	log.Printf("Creating variants for all %d products...", len(productsNeedingVariants))

	variantCount := 0
	for i, product := range productsNeedingVariants {
		// Get category name
		categoryName := ""
		if product.Edges.Category != nil {
			categoryName = product.Edges.Category.Name
		}

		// Generate variants based on category
		variants := generateVariantsForProduct(client, product, categoryName)

		// Create variants
		for _, variantBuilder := range variants {
			_, err := variantBuilder.Save(ctx)
			if err != nil {
				log.Printf("Warning: failed to create variant for product %s: %v", product.Slug, err)
				continue
			}
			variantCount++
		}

		// Log progress every 10%
		if (i+1)%((len(productsNeedingVariants))/10) == 0 || i == len(productsNeedingVariants)-1 {
			log.Printf("Variant creation progress: %d/%d products (%.1f%%), %d variants created",
				i+1, len(productsNeedingVariants), float64(i+1)/float64(len(productsNeedingVariants))*100, variantCount)
		}
	}

	variantDuration := time.Since(variantStartTime)
	log.Printf("Successfully created %d variants in %v", variantCount, variantDuration)
	log.Printf("Average: %.2f variants/second", float64(variantCount)/variantDuration.Seconds())
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
	for _, brandName := range fashionBrands {
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
		"Dresses":         fmt.Sprintf("Elegant dress from %s crafted with premium fabric for a flawless fit. Perfect for special occasions or everyday elegance.", brand),
		"Tops & Blouses":  fmt.Sprintf("Stylish top from %s designed for comfort and versatility. Essential piece for any modern wardrobe.", brand),
		"Pants & Jeans":   fmt.Sprintf("Premium denim from %s with superior stretch and durability. Classic fit that never goes out of style.", brand),
		"Skirts":          fmt.Sprintf("Chic skirt from %s featuring elegant design and comfortable fit. Perfect for both casual and formal occasions.", brand),
		"Jackets & Coats": fmt.Sprintf("Stylish outerwear from %s combining fashion and function. Premium materials provide warmth and sophistication.", brand),
		"Knitwear":        fmt.Sprintf("Cozy knitwear from %s made from soft, premium materials. Perfect for layering in any season.", brand),
		"Sneakers":        fmt.Sprintf("Trendy sneakers from %s with superior comfort and style. Perfect for urban adventures and casual outings.", brand),
		"Boots":          fmt.Sprintf("Classic boots from %s crafted with quality leather for timeless style and durability.", brand),
		"Heels":          fmt.Sprintf("Elegant heels from %s designed for sophistication and comfort. Perfect statement piece for any outfit.", brand),
		"Bags & Handbags": fmt.Sprintf("Luxurious handbag from %s featuring exquisite craftsmanship and elegant design. The perfect accessory for any occasion.", brand),
		"Watches":        fmt.Sprintf("Elegant timepiece from %s combining precision engineering with sophisticated design. A statement of refined taste.", brand),
		"Sunglasses":     fmt.Sprintf("Stylish sunglasses from %s with UV protection and contemporary design. Essential accessory for sunny days.", brand),
		"Athletic Wear":  fmt.Sprintf("Performance athletic wear from %s designed for comfort and movement. Perfect for workouts and active lifestyles.", brand),
		"Swimwear":       fmt.Sprintf("Stylish swimwear from %s with premium fabric that's chlorine-resistant. Perfect for beach or pool days.", brand),
	}

	if desc, exists := descriptions[category]; exists {
		return desc
	}

	// Default description
	return fmt.Sprintf("Quality fashion item from %s designed for style and comfort. Crafted with premium materials for everyday wear.", brand)
}

// generateRealisticPrice generates realistic pricing based on category
func generateRealisticPrice(category string) float64 {
	priceRanges := map[string]struct {
		min, max float64
	}{
		"Dresses":         {50, 300},
		"Tops & Blouses":  {20, 150},
		"Pants & Jeans":   {30, 200},
		"Skirts":          {25, 180},
		"Jackets & Coats": {80, 500},
		"Knitwear":        {40, 250},
		"Sneakers":        {60, 300},
		"Boots":           {80, 400},
		"Heels":           {70, 350},
		"Flats":           {40, 200},
		"Sandals":         {30, 150},
		"Bags & Handbags": {50, 500},
		"Wallets":         {20, 150},
		"Belts":           {15, 100},
		"Watches":         {100, 1000},
		"Sunglasses":      {30, 300},
		"Athletic Wear":   {25, 150},
		"Swimwear":        {30, 180},
		"Lingerie":        {20, 120},
		"Pajamas":         {25, 100},
		"Robes":           {40, 150},
		"Shorts":          {20, 100},
		"Shirts":          {30, 180},
		"T-Shirts & Polos": {15, 80},
		"Suits & Blazers": {150, 800},
		"Underwear":       {10, 50},
	}

	if prange, exists := priceRanges[category]; exists {
		return float64(rand.IntN(int(prange.max-prange.min)*100)+int(prange.min)*100) / 100.0
	}

	// Default price range
	return float64(rand.IntN(150)+20) + float64(rand.IntN(100))/100.0
}

// generateRealisticDimensions generates realistic dimensions based on category
func generateRealisticDimensions(category string) (weight, length, width, height int) {
	switch {
	case strings.Contains(category, "Dress") || strings.Contains(category, "Coat") || strings.Contains(category, "Jacket"):
		weight = rand.IntN(800) + 300         // 300-1100 g
		length = rand.IntN(20) + 60           // 60-80 cm
		width = rand.IntN(20) + 40            // 40-60 cm
		height = rand.IntN(10) + 5            // 5-15 cm

	case strings.Contains(category, "Pants") || strings.Contains(category, "Jeans"):
		weight = rand.IntN(400) + 200         // 200-600 g
		length = rand.IntN(20) + 100          // 100-120 cm
		width = rand.IntN(15) + 30            // 30-45 cm
		height = rand.IntN(5) + 3             // 3-8 cm

	case strings.Contains(category, "Top") || strings.Contains(category, "Shirt") || strings.Contains(category, "Blouse"):
		weight = rand.IntN(300) + 100         // 100-400 g
		length = rand.IntN(15) + 60           // 60-75 cm
		width = rand.IntN(15) + 40            // 40-55 cm
		height = rand.IntN(5) + 2             // 2-7 cm

	case strings.Contains(category, "Skirt"):
		weight = rand.IntN(300) + 100         // 100-400 g
		length = rand.IntN(30) + 40           // 40-70 cm
		width = rand.IntN(20) + 35            // 35-55 cm
		height = rand.IntN(5) + 3             // 3-8 cm

	case strings.Contains(category, "Sneaker") || strings.Contains(category, "Boot") || strings.Contains(category, "Shoe"):
		weight = rand.IntN(500) + 300         // 300-800 g (per shoe)
		length = rand.IntN(10) + 30           // 30-40 cm
		width = rand.IntN(5) + 12             // 12-17 cm
		height = rand.IntN(15) + 10           // 10-25 cm

	case strings.Contains(category, "Bag") || strings.Contains(category, "Handbag"):
		weight = rand.IntN(800) + 200         // 200-1000 g
		length = rand.IntN(20) + 25           // 25-45 cm
		width = rand.IntN(10) + 15            // 15-25 cm
		height = rand.IntN(15) + 20           // 20-35 cm

	case strings.Contains(category, "Watch"):
		weight = rand.IntN(100) + 50          // 50-150 g
		length = rand.IntN(5) + 20            // 20-25 cm
		width = rand.IntN(3) + 3              // 3-6 cm
		height = rand.IntN(5) + 5             // 5-10 cm

	case strings.Contains(category, "Sunglasses"):
		weight = rand.IntN(50) + 20           // 20-70 g
		length = rand.IntN(5) + 15            // 15-20 cm
		width = rand.IntN(3) + 5              // 5-8 cm
		height = rand.IntN(3) + 5             // 5-8 cm

	case strings.Contains(category, "Belt"):
		weight = rand.IntN(100) + 50          // 50-150 g
		length = rand.IntN(20) + 100          // 100-120 cm
		width = rand.IntN(5) + 5              // 5-10 cm
		height = rand.IntN(3) + 1             // 1-4 cm

	default:
		// Default dimensions for fashion items
		weight = rand.IntN(500) + 100         // 100-600 g
		length = rand.IntN(30) + 20           // 20-50 cm
		width = rand.IntN(20) + 15            // 15-35 cm
		height = rand.IntN(10) + 5            // 5-15 cm
	}

	return weight, length, width, height
}

func createProductBuilder(client *ent.Client, num int, brandMap map[string]uuid.UUID, categoryMap map[string]uuid.UUID) *ent.ProductCreate {
	// Select a random leaf category for product assignment
	leafCategories := getLeafCategories()
	selectedCategory := leafCategories[rand.IntN(len(leafCategories))]
	categoryID := categoryMap[selectedCategory]

	// Select a random brand
	selectedBrand := fashionBrands[rand.IntN(len(fashionBrands))]
	brandID := brandMap[selectedBrand]

	// Generate product name using templates or fallback
	productName := generateProductName(selectedBrand, selectedCategory)
	slug := fmt.Sprintf("%s-%d", strings.ToLower(strings.ReplaceAll(productName, " ", "-")), num)

	// Generate realistic price based on category (now called basePrice)
	basePrice := generateRealisticPrice(selectedCategory)

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

	builder := client.Product.Create().
		SetSlug(slug).
		SetName(productName).
		SetBasePrice(basePrice).
		SetDescription(description).
		SetLowStockThreshold(10). // Default threshold
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

	return builder
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

// Variant attribute combinations by category
var variantAttributesByCategory = map[string][]map[string]string{
	"Dresses": {
		{"size": "XS", "color": "Black"},
		{"size": "XS", "color": "White"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "White"},
		{"size": "S", "color": "Red"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "White"},
		{"size": "M", "color": "Red"},
		{"size": "M", "color": "Blue"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "White"},
		{"size": "L", "color": "Blue"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "White"},
	},
	"Tops & Blouses": {
		{"size": "XS", "color": "White"},
		{"size": "XS", "color": "Black"},
		{"size": "S", "color": "White"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Pink"},
		{"size": "M", "color": "White"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Pink"},
		{"size": "M", "color": "Blue"},
		{"size": "L", "color": "White"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Blue"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "White"},
	},
	"Pants & Jeans": {
		{"size": "26", "color": "Blue"},
		{"size": "26", "color": "Black"},
		{"size": "28", "color": "Blue"},
		{"size": "28", "color": "Black"},
		{"size": "30", "color": "Blue"},
		{"size": "30", "color": "Black"},
		{"size": "32", "color": "Blue"},
		{"size": "32", "color": "Black"},
		{"size": "34", "color": "Blue"},
		{"size": "34", "color": "Black"},
		{"size": "36", "color": "Blue"},
		{"size": "36", "color": "Black"},
	},
	"Skirts": {
		{"size": "XS", "color": "Black"},
		{"size": "XS", "color": "Red"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Red"},
		{"size": "S", "color": "Beige"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Red"},
		{"size": "M", "color": "Beige"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Beige"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "Beige"},
	},
	"Jackets & Coats": {
		{"size": "XS", "color": "Black"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Brown"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Brown"},
		{"size": "M", "color": "Beige"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Brown"},
		{"size": "L", "color": "Beige"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "Brown"},
	},
	"Knitwear": {
		{"size": "XS", "color": "Gray"},
		{"size": "S", "color": "Gray"},
		{"size": "S", "color": "Navy"},
		{"size": "M", "color": "Gray"},
		{"size": "M", "color": "Navy"},
		{"size": "M", "color": "White"},
		{"size": "L", "color": "Gray"},
		{"size": "L", "color": "Navy"},
		{"size": "L", "color": "White"},
		{"size": "XL", "color": "Navy"},
		{"size": "XL", "color": "Gray"},
	},
	"Sneakers": {
		{"size": "36", "color": "White"},
		{"size": "37", "color": "White"},
		{"size": "37", "color": "Black"},
		{"size": "38", "color": "White"},
		{"size": "38", "color": "Black"},
		{"size": "39", "color": "White"},
		{"size": "39", "color": "Black"},
		{"size": "40", "color": "White"},
		{"size": "40", "color": "Black"},
		{"size": "41", "color": "White"},
		{"size": "41", "color": "Black"},
		{"size": "42", "color": "White"},
		{"size": "42", "color": "Black"},
		{"size": "43", "color": "White"},
		{"size": "43", "color": "Black"},
		{"size": "44", "color": "Black"},
		{"size": "45", "color": "Black"},
	},
	"Boots": {
		{"size": "36", "color": "Black"},
		{"size": "37", "color": "Black"},
		{"size": "37", "color": "Brown"},
		{"size": "38", "color": "Black"},
		{"size": "38", "color": "Brown"},
		{"size": "39", "color": "Black"},
		{"size": "39", "color": "Brown"},
		{"size": "40", "color": "Black"},
		{"size": "40", "color": "Brown"},
		{"size": "41", "color": "Black"},
		{"size": "41", "color": "Brown"},
		{"size": "42", "color": "Black"},
		{"size": "42", "color": "Brown"},
		{"size": "43", "color": "Black"},
		{"size": "43", "color": "Brown"},
		{"size": "44", "color": "Black"},
		{"size": "45", "color": "Black"},
	},
	"Heels": {
		{"size": "36", "color": "Black"},
		{"size": "36", "color": "Red"},
		{"size": "37", "color": "Black"},
		{"size": "37", "color": "Red"},
		{"size": "37", "color": "Nude"},
		{"size": "38", "color": "Black"},
		{"size": "38", "color": "Red"},
		{"size": "38", "color": "Nude"},
		{"size": "39", "color": "Black"},
		{"size": "39", "color": "Red"},
		{"size": "39", "color": "Nude"},
		{"size": "40", "color": "Black"},
		{"size": "40", "color": "Red"},
		{"size": "41", "color": "Black"},
	},
	"Bags & Handbags": {
		{"size": "Small", "color": "Black"},
		{"size": "Small", "color": "Brown"},
		{"size": "Small", "color": "Beige"},
		{"size": "Medium", "color": "Black"},
		{"size": "Medium", "color": "Brown"},
		{"size": "Medium", "color": "Beige"},
		{"size": "Medium", "color": "Red"},
		{"size": "Large", "color": "Black"},
		{"size": "Large", "color": "Brown"},
		{"size": "Large", "color": "Beige"},
	},
	"Watches": {
		{"size": "Small", "color": "Silver"},
		{"size": "Small", "color": "Gold"},
		{"size": "Small", "color": "Rose Gold"},
		{"size": "Medium", "color": "Silver"},
		{"size": "Medium", "color": "Gold"},
		{"size": "Medium", "color": "Rose Gold"},
		{"size": "Large", "color": "Silver"},
		{"size": "Large", "color": "Gold"},
		{"size": "Large", "color": "Black"},
	},
	"Sunglasses": {
		{"color": "Black"},
		{"color": "Brown"},
		{"color": "Tortoise"},
		{"color": "Gradient"},
		{"color": "Mirrored"},
	},
	"Athletic Wear": {
		{"size": "XS", "color": "Black"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Navy"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Navy"},
		{"size": "M", "color": "Gray"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Navy"},
		{"size": "L", "color": "Gray"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "Navy"},
	},
	"Swimwear": {
		{"size": "XS", "color": "Black"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Blue"},
		{"size": "S", "color": "Red"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Blue"},
		{"size": "M", "color": "Red"},
		{"size": "M", "color": "Floral"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Blue"},
		{"size": "L", "color": "Floral"},
		{"size": "XL", "color": "Black"},
	},
	"Shirts": {
		{"size": "S", "color": "White"},
		{"size": "S", "color": "Blue"},
		{"size": "M", "color": "White"},
		{"size": "M", "color": "Blue"},
		{"size": "M", "color": "Pink"},
		{"size": "L", "color": "White"},
		{"size": "L", "color": "Blue"},
		{"size": "L", "color": "Pink"},
		{"size": "XL", "color": "White"},
		{"size": "XL", "color": "Blue"},
	},
	"T-Shirts & Polos": {
		{"size": "XS", "color": "White"},
		{"size": "S", "color": "White"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Navy"},
		{"size": "M", "color": "White"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Navy"},
		{"size": "M", "color": "Gray"},
		{"size": "L", "color": "White"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Navy"},
		{"size": "L", "color": "Gray"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "Navy"},
	},
	"Shorts": {
		{"size": "XS", "color": "Black"},
		{"size": "S", "color": "Black"},
		{"size": "S", "color": "Navy"},
		{"size": "M", "color": "Black"},
		{"size": "M", "color": "Navy"},
		{"size": "M", "color": "Beige"},
		{"size": "L", "color": "Black"},
		{"size": "L", "color": "Navy"},
		{"size": "L", "color": "Beige"},
		{"size": "XL", "color": "Black"},
		{"size": "XL", "color": "Navy"},
	},
}

// generateVariantsForProduct generates variant builders for a product based on its category
func generateVariantsForProduct(client *ent.Client, product *ent.Product, categoryName string) []*ent.ProductVariantCreate {
	var variants []*ent.ProductVariantCreate

	// Get variant templates for category
	templates, hasVariants := variantAttributesByCategory[categoryName]
	if !hasVariants {
		// Default: create 2-3 generic variants
		for i := 0; i < rand.IntN(2)+2; i++ {
			variantNum := i + 1
			attributes := map[string]string{
				"variant": fmt.Sprintf("Option %d", variantNum),
			}

			variants = append(variants, createVariantBuilder(client, product, variantNum, attributes, 0))
		}
		return variants
	}

	// Create variants from templates
	for i, attributes := range templates {
		variantNum := i + 1

		// Calculate price adjustment based on attributes
		priceAdjustment := calculatePriceAdjustment(categoryName, attributes)

		variants = append(variants, createVariantBuilder(client, product, variantNum, attributes, priceAdjustment))
	}

	return variants
}

// createVariantBuilder creates a variant builder with realistic data
func createVariantBuilder(client *ent.Client, product *ent.Product, variantNum int, attributes map[string]string, priceAdjustment float64) *ent.ProductVariantCreate {
	// Generate SKU
	sku := fmt.Sprintf("%s-VAR%d", strings.ToUpper(product.Slug[:min(10, len(product.Slug))]), variantNum)

	// Random stock quantity
	stockQuantity := rand.IntN(100) + 1 // 1-100 units

	// 90% active
	isActive := rand.Float64() < 0.9

	// Random created_at in the past (after product creation)
	createdAt := product.CreatedAt.Add(time.Duration(rand.IntN(30)) * 24 * time.Hour)
	updatedAt := createdAt.Add(time.Duration(rand.IntN(30)) * 24 * time.Hour)

	return client.ProductVariant.Create().
		SetSku(sku).
		SetAttributes(attributes).
		SetStockQuantity(stockQuantity).
		SetPriceAdjustment(priceAdjustment).
		SetIsActive(isActive).
		SetProductID(product.ID).
		SetCreatedAt(createdAt).
		SetUpdatedAt(updatedAt)
}

// calculatePriceAdjustment calculates price adjustment based on variant attributes
func calculatePriceAdjustment(category string, attributes map[string]string) float64 {
	adjustment := 0.0

	// Size-based adjustments (larger sizes cost more)
	if size, ok := attributes["size"]; ok {
		switch {
		case strings.Contains(size, "XL") || strings.Contains(size, "44") || strings.Contains(size, "45") || strings.Contains(size, "36") || size == "Large":
			adjustment += 15.0
		case strings.Contains(size, "L") || strings.Contains(size, "42") || strings.Contains(size, "43") || strings.Contains(size, "34") || size == "Medium":
			adjustment += 10.0
		case strings.Contains(size, "M") || strings.Contains(size, "40") || strings.Contains(size, "41") || strings.Contains(size, "32") || strings.Contains(size, "30"):
			adjustment += 5.0
		}
	}

	// Color-based adjustments (premium colors cost more)
	if color, ok := attributes["color"]; ok {
		switch color {
		case "Gold", "Rose Gold", "Silver", "Black", "Nude":
			adjustment += 10.0
		case "Red", "Blue", "Pink", "Brown":
			adjustment += 5.0
		case "Beige", "Gray", "White", "Navy":
			adjustment += 0.0
		}
	}

	// Material-based adjustments
	if material, ok := attributes["material"]; ok {
		switch material {
		case "Leather", "Genuine Leather":
			adjustment += 50.0
		case "Suede", "Velvet":
			adjustment += 30.0
		case "Silk", "Cashmere":
			adjustment += 40.0
		case "Cotton", "Linen":
			adjustment += 10.0
		}
	}

	return adjustment
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createAdminUser creates a default admin user if it doesn't exist
func createAdminUser(ctx context.Context, client *ent.Client) error {
	// Check if admin user already exists
	existingUser, err := client.User.Query().
		Where(user.EmailEQ("admin")).
		Only(ctx)

	if err == nil && existingUser != nil {
		log.Printf("Admin user already exists with email: %s", existingUser.Email)
		return nil
	}

	// Hash the password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("adminadmin"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := &entities.User{
		Email:        "admin",
		PasswordHash: string(passwordHash),
		Name:         "Admin",
		Role:         entities.UserRoleAdmin,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Convert to ent user creation
	_, err = client.User.Create().
		SetEmail(adminUser.Email).
		SetPasswordHash(adminUser.PasswordHash).
		SetName(adminUser.Name).
		SetRole(user.Role(adminUser.Role)).
		SetIsActive(adminUser.IsActive).
		SetCreatedAt(adminUser.CreatedAt).
		SetUpdatedAt(adminUser.UpdatedAt).
		Save(ctx)

	if err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("Admin user created successfully with email: %s", adminUser.Email)
	return nil
}
