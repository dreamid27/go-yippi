package rajaongkir

// Province represents a RajaOngkir province
type Province struct {
	ProvinceID string `json:"province_id"`
	Province   string `json:"province"`
}

// City represents a RajaOngkir city
type City struct {
	CityID     string `json:"city_id"`
	ProvinceID string `json:"province_id"`
	Province   string `json:"province"`
	Type       string `json:"type"`         // "Kota" or "Kabupaten"
	CityName   string `json:"city_name"`
	PostalCode string `json:"postal_code"`
}

// CostRequest represents a RajaOngkir cost calculation request
type CostRequest struct {
	Origin      string `json:"origin"`       // city ID
	Destination string `json:"destination"`  // city ID
	Weight      int    `json:"weight"`       // in grams
	Courier     string `json:"courier"`      // "jne", "tiki", "pos", etc.
}

// CostResponse represents a RajaOngkir cost calculation response
type CostResponse struct {
	RajaOngkir struct {
		OriginDetails      City             `json:"origin_details"`
		DestinationDetails City             `json:"destination_details"`
		Results            []CourierResults `json:"results"`
	} `json:"rajaongkir"`
}

// CourierResults represents cost results for a single courier
type CourierResults struct {
	Code   string    `json:"code"`   // "jne", "tiki", etc.
	Name   string    `json:"name"`   // "JNE", "TIKI", etc.
	Costs  []CostItem `json:"costs"`
}

// CostItem represents a single shipping cost option
type CostItem struct {
	Service     string        `json:"service"`  // "OKE", "REG", "YES", etc.
	Description string        `json:"description"`
	Cost        []CostDetail `json:"cost"`
}

// CostDetail represents cost breakdown
type CostDetail struct {
	Value int    `json:"value"` // in IDR
	Etd   string `json:"etd"`   // estimated delivery time, e.g., "2-3 days"
	Note  string `json:"note"`
}

// ProvincesResponse represents RajaOngkir provinces list response
type ProvincesResponse struct {
	RajaOngkir struct {
		Province []Province `json:"province"`
	} `json:"rajaongkir"`
}

// CitiesResponse represents RajaOngkir cities list response
type CitiesResponse struct {
	RajaOngkir struct {
		City []City `json:"city"`
	} `json:"rajaongkir"`
}