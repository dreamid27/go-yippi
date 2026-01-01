package rajaongkir

import (
	"testing"
	"time"
)

func TestCache_SetAndGetProvinces(t *testing.T) {
	cache := NewCache()

	provinces := []Province{
		{ProvinceID: "1", Province: "Jawa Barat"},
		{ProvinceID: "2", Province: "DKI Jakarta"},
	}

	cache.SetProvinces(provinces)

	result, found := cache.GetProvinces()
	if !found {
		t.Fatal("expected to find provinces in cache")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 provinces, got %d", len(result))
	}

	// Verify province data
	if result[0].ProvinceID != "1" || result[0].Province != "Jawa Barat" {
		t.Errorf("expected first province to be '1', 'Jawa Barat', got '%s', '%s'",
			result[0].ProvinceID, result[0].Province)
	}

	if result[1].ProvinceID != "2" || result[1].Province != "DKI Jakarta" {
		t.Errorf("expected second province to be '2', 'DKI Jakarta', got '%s', '%s'",
			result[1].ProvinceID, result[1].Province)
	}
}

func TestCache_ProvincesExpiry(t *testing.T) {
	cache := NewCache()

	provinces := []Province{
		{ProvinceID: "1", Province: "Jawa Barat"},
	}

	cache.SetProvinces(provinces)

	// Test before expiry
	result, found := cache.GetProvinces()
	if !found {
		t.Error("expected to find provinces before expiry")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 province, got %d", len(result))
	}

	// Manually set expiry to past time
	cache.SetProvinces([]Province{}) // Clear first
	cache.mu.Lock()
	cache.provincesExpiry = time.Now().Add(-1 * time.Hour)
	cache.mu.Unlock()

	// Test after expiry
	_, found = cache.GetProvinces()
	if found {
		t.Error("expected provinces to be expired and not found")
	}
}

func TestCache_SetAndGetCities(t *testing.T) {
	cache := NewCache()

	cities := []City{
		{CityID: "1", CityName: "Bandung", ProvinceID: "1", Province: "Jawa Barat", Type: "Kota", PostalCode: "40111"},
		{CityID: "2", CityName: "Jakarta", ProvinceID: "2", Province: "DKI Jakarta", Type: "Kota", PostalCode: "11001"},
	}

	cache.SetCities(cities)

	result, found := cache.GetCities()
	if !found {
		t.Fatal("expected to find cities in cache")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 cities, got %d", len(result))
	}

	// Verify city data (order not guaranteed since we're iterating a map)
	foundBandung := false
	foundJakarta := false

	for _, city := range result {
		if city.CityName == "Bandung" {
			foundBandung = true
			if city.Province != "Jawa Barat" {
				t.Errorf("expected Bandung to be in 'Jawa Barat', got '%s'", city.Province)
			}
		} else if city.CityName == "Jakarta" {
			foundJakarta = true
			if city.Province != "DKI Jakarta" {
				t.Errorf("expected Jakarta to be in 'DKI Jakarta', got '%s'", city.Province)
			}
		}
	}

	if !foundBandung {
		t.Error("expected to find Bandung in cities")
	}
	if !foundJakarta {
		t.Error("expected to find Jakarta in cities")
	}
}

func TestCache_GetCity(t *testing.T) {
	cache := NewCache()

	cities := []City{
		{CityID: "1", CityName: "Bandung", ProvinceID: "1", Province: "Jawa Barat", Type: "Kota", PostalCode: "40111"},
		{CityID: "2", CityName: "Jakarta", ProvinceID: "2", Province: "DKI Jakarta", Type: "Kota", PostalCode: "11001"},
	}

	cache.SetCities(cities)

	// Test existing city
	city, found := cache.GetCity("1")
	if !found {
		t.Fatal("expected to find city '1' in cache")
	}

	if city.CityName != "Bandung" {
		t.Errorf("expected city name 'Bandung', got '%s'", city.CityName)
	}

	if city.ProvinceID != "1" {
		t.Errorf("expected province ID '1', got '%s'", city.ProvinceID)
	}

	// Test non-existing city
	_, found = cache.GetCity("999")
	if found {
		t.Error("expected city '999' to not exist in cache")
	}

	// Test with expired cache
	cache.SetCities([]City{}) // Clear first
	cache.mu.Lock()
	cache.citiesExpiry = time.Now().Add(-1 * time.Hour)
	cache.mu.Unlock()

	_, found = cache.GetCity("1")
	if found {
		t.Error("expected city to not be found when cache is expired")
	}
}

func TestCache_EmptyCache(t *testing.T) {
	cache := NewCache()

	// Test empty cache for provinces
	_, found := cache.GetProvinces()
	if found {
		t.Error("expected no provinces in empty cache")
	}

	// Test empty cache for cities
	_, found = cache.GetCities()
	if found {
		t.Error("expected no cities in empty cache")
	}

	// Test get city from empty cache
	city, found := cache.GetCity("1")
	if found {
		t.Error("expected no city '1' in empty cache")
	}

	if city != nil {
		t.Error("expected nil city when not found")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	cache := NewCache()

	// Initialize data
	provinces := []Province{
		{ProvinceID: "1", Province: "Jawa Barat"},
		{ProvinceID: "2", Province: "DKI Jakarta"},
	}

	// Test concurrent writes and reads
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(p []Province) {
			cache.SetProvinces(p)
			done <- true
		}(provinces)
	}

	for i := 0; i < 10; i++ {
		go func() {
			_, _ = cache.GetProvinces()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify data is still consistent
	result, found := cache.GetProvinces()
	if !found {
		t.Fatal("expected to find provinces after concurrent access")
	}

	if len(result) != 2 {
		t.Errorf("expected 2 provinces after concurrent access, got %d", len(result))
	}
}