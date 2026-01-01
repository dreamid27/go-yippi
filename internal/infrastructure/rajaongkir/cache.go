package rajaongkir

import (
	"sync"
	"time"
)

// Cache represents an in-memory cache for RajaOngkir data
type Cache struct {
	mu                sync.RWMutex
	provinces         map[string]Province // key: provinceID
	cities            map[string]City    // key: cityID
	provincesExpiry   time.Time
	citiesExpiry      time.Time
}

// NewCache creates a new RajaOngkir cache instance
func NewCache() *Cache {
	return &Cache{
		provinces: make(map[string]Province),
		cities:    make(map[string]City),
	}
}

// SetProvinces caches provinces with 24-hour TTL
func (c *Cache) SetProvinces(provinces []Province) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.provinces = make(map[string]Province, len(provinces))
	for _, province := range provinces {
		c.provinces[province.ProvinceID] = province
	}
	c.provincesExpiry = time.Now().Add(24 * time.Hour)
}

// GetProvinces retrieves cached provinces if not expired
func (c *Cache) GetProvinces() ([]Province, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().After(c.provincesExpiry) {
		return nil, false
	}

	result := make([]Province, 0, len(c.provinces))
	for _, province := range c.provinces {
		result = append(result, province)
	}
	return result, true
}

// SetCities caches cities with 24-hour TTL
func (c *Cache) SetCities(cities []City) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cities = make(map[string]City, len(cities))
	for _, city := range cities {
		c.cities[city.CityID] = city
	}
	c.citiesExpiry = time.Now().Add(24 * time.Hour)
}

// GetCities retrieves cached cities if not expired
func (c *Cache) GetCities() ([]City, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().After(c.citiesExpiry) {
		return nil, false
	}

	result := make([]City, 0, len(c.cities))
	for _, city := range c.cities {
		result = append(result, city)
	}
	return result, true
}

// GetCity retrieves a single city by ID
func (c *Cache) GetCity(cityID string) (*City, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if time.Now().After(c.citiesExpiry) {
		return nil, false
	}

	city, exists := c.cities[cityID]
	if !exists {
		return nil, false
	}
	return &city, true
}