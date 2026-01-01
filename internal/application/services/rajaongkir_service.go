package services

import (
	"context"
	"fmt"
	"example.com/go-yippi/internal/infrastructure/rajaongkir"
)

// RajaOngkirService handles RajaOngkir API operations
type RajaOngkirService struct {
	client *rajaongkir.Client
	cache  *rajaongkir.Cache
}

// NewRajaOngkirService creates a new RajaOngkir service instance
func NewRajaOngkirService(client *rajaongkir.Client, cache *rajaongkir.Cache) *RajaOngkirService {
	return &RajaOngkirService{
		client: client,
		cache:  cache,
	}
}

// GetProvinces retrieves all provinces, using cache if available
func (s *RajaOngkirService) GetProvinces(ctx context.Context) ([]rajaongkir.Province, error) {
	// Try cache first
	provinces, found := s.cache.GetProvinces()
	if found {
		return provinces, nil
	}

	// Cache miss - fetch from API
	resp, err := s.client.GetProvinces()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch provinces: %w", err)
	}

	provinces = resp.RajaOngkir.Province

	// Update cache
	s.cache.SetProvinces(provinces)

	return provinces, nil
}

// GetCities retrieves all cities, using cache if available
func (s *RajaOngkirService) GetCities(ctx context.Context) ([]rajaongkir.City, error) {
	// Try cache first
	cities, found := s.cache.GetCities()
	if found {
		return cities, nil
	}

	// Cache miss - fetch from API
	resp, err := s.client.GetCities()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cities: %w", err)
	}

	cities = resp.RajaOngkir.City

	// Update cache
	s.cache.SetCities(cities)

	return cities, nil
}

// GetCity retrieves a single city by ID, using cache if available
func (s *RajaOngkirService) GetCity(ctx context.Context, cityID string) (*rajaongkir.City, error) {
	// Try cache first
	city, found := s.cache.GetCity(cityID)
	if found {
		return city, nil
	}

	// Cache miss - fetch all cities from API
	cities, err := s.GetCities(ctx)
	if err != nil {
		return nil, err
	}

	// Find the specific city
	for _, c := range cities {
		if c.CityID == cityID {
			return &c, nil
		}
	}

	return nil, fmt.Errorf("city with ID %s not found", cityID)
}