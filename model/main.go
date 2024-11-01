package model

// ASN represents the ASN database structure.
type ASN struct {
	Number       uint   `maxminddb:"autonomous_system_number"`
	Organization string `maxminddb:"autonomous_system_organization"`
}

// Location represents the geographical location with latitude and longitude.
type Location struct {
	Latitude  float64 `maxminddb:"latitude"`
	Longitude float64 `maxminddb:"longitude"`
}

// City represents the City database structure.
type City struct {
	Country           CountryInfo   `maxminddb:"country"`
	RegisteredCountry CountryInfo   `maxminddb:"registered_country"`
	Subdivisions      []Subdivision `maxminddb:"subdivisions"`
	City              NameInfo      `maxminddb:"city"`
	Location          Location      `maxminddb:"location"` // Add this line
}

// CountryInfo represents country information in the database.
type CountryInfo struct {
	ISOCode string            `maxminddb:"iso_code"`
	Names   map[string]string `maxminddb:"names"`
}

// Subdivision represents subdivision information in the database.
type Subdivision struct {
	Names map[string]string `maxminddb:"names"`
}

// NameInfo represents name information in the database.
type NameInfo struct {
	Names map[string]string `maxminddb:"names"`
}

// GeoCN represents the GeoCN database structure.
type GeoCN struct {
	Province  string `maxminddb:"province"`
	City      string `maxminddb:"city"`
	Districts string `maxminddb:"districts"`
	ISP       string `maxminddb:"isp"`
	Net       string `maxminddb:"net"`
}
