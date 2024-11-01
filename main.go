package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// ASN represents the ASN database structure.
type ASN struct {
	Number       uint   `maxminddb:"autonomous_system_number"`
	Organization string `maxminddb:"autonomous_system_organization"`
}

// City represents the City database structure.
type City struct {
	Country           CountryInfo   `maxminddb:"country"`
	RegisteredCountry CountryInfo   `maxminddb:"registered_country"`
	Subdivisions      []Subdivision `maxminddb:"subdivisions"`
	City              NameInfo      `maxminddb:"city"`
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

var (
	cityReader, asnReader, cnReader *maxminddb.Reader
	asnMap                          = map[uint]string{
		9812: "东方有线", 9389: "中国长城", 17962: "天威视讯",
		// Add more mappings as needed
	}
)

func init() {
	var err error
	//cityReader, err = maxminddb.Open(os.Getenv("CITY_DB_PATH"))
	cityReader, err = maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalf("Error opening city database: %v", err)
	}
	//asnReader, err = maxminddb.Open(os.Getenv("ASN_DB_PATH"))
	asnReader, err = maxminddb.Open("GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatalf("Error opening ASN database: %v", err)
	}
	//cnReader, err = maxminddb.Open(os.Getenv("CN_DB_PATH"))
	cnReader, err = maxminddb.Open("GeoCN.mmdb")
	if err != nil {
		log.Fatalf("Error opening CN database: %v", err)
	}
}

func getAsInfo(number uint) string {
	return asnMap[number]
}

func getDescription(names map[string]string) string {
	if name, ok := names["zh-CN"]; ok {
		return name
	}
	return names["en"]
}

func getCountry(d map[string]string) string {
	r := getDescription(d)
	switch r {
	case "香港", "澳门", "台湾":
		return "中国" + r
	default:
		return r
	}
}

func deDuplicate(regions []string) []string {
	unique := make(map[string]bool)
	var result []string
	for _, region := range regions {
		if _, ok := unique[region]; !ok && region != "" {
			unique[region] = true
			result = append(result, region)
		}
	}
	return result
}

func getMaxmind(ip string) (map[string]interface{}, error) {
	ret := make(map[string]interface{})
	ret["ip"] = ip

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	var asn ASN
	if err := asnReader.Lookup(parsedIP, &asn); err == nil {
		asInfo := map[string]interface{}{
			"number": asn.Number,
			"name":   asn.Organization,
		}
		if info := getAsInfo(asn.Number); info != "" {
			asInfo["info"] = info
		}
		ret["as"] = asInfo
	}

	var city City
	if network, ok, err := cityReader.LookupNetwork(parsedIP, &city); err == nil && ok {
		ret["addr"] = network.String()
		if city.Country.ISOCode != "" {
			ret["country"] = map[string]string{
				"code": city.Country.ISOCode,
				"name": getCountry(city.Country.Names),
			}
		}
		if city.RegisteredCountry.ISOCode != "" {
			ret["registered_country"] = map[string]string{
				"code": city.RegisteredCountry.ISOCode,
				"name": getCountry(city.RegisteredCountry.Names),
			}
		}
		var regions []string
		for _, subdivision := range city.Subdivisions {
			regions = append(regions, getDescription(subdivision.Names))
		}
		if cityName := getDescription(city.City.Names); cityName != "" {
			regions = append(regions, cityName)
		}
		ret["regions"] = deDuplicate(regions)
	}

	return ret, nil
}

func getCn(ip string, info map[string]interface{}) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return
	}

	var geoCN GeoCN
	if network, ok, err := cnReader.LookupNetwork(parsedIP, &geoCN); err == nil && ok {
		info["addr"] = network.String()
		regions := deDuplicate([]string{geoCN.Province, geoCN.City, geoCN.Districts})
		if len(regions) > 0 {
			info["regions"] = regions
		}
		if _, ok := info["as"]; !ok {
			info["as"] = make(map[string]interface{})
		}
		info["as"].(map[string]interface{})["info"] = geoCN.ISP
		if geoCN.Net != "" {
			info["type"] = geoCN.Net
		}
	}
}

func getIpInfo(ip string) map[string]interface{} {
	info, err := getMaxmind(ip)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}
	if country, ok := info["country"].(map[string]string); ok && country["code"] == "CN" {
		if regCountry, ok := info["registered_country"].(map[string]string); !ok || regCountry["code"] == "CN" {
			getCn(ip, info)
		}
	}
	return info
}

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		ip := c.Query("ip")
		if ip == "" {
			ip = c.ClientIP()
		}
		info := getIpInfo(ip)
		c.JSON(http.StatusOK, info)
	})

	r.GET("/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		info := getIpInfo(ip)
		c.JSON(http.StatusOK, info)
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}

func init() {
	loadDatabases()
	// 监听SIGHUP信号以重新加载数据库
	go func() {
		sighup := make(chan os.Signal, 1)
		signal.Notify(sighup, syscall.SIGHUP)
		for {
			<-sighup
			log.Println("Received SIGHUP, reloading databases...")
			loadDatabases()
		}
	}()
}

func loadDatabases() {
	var err error
	cityReader, err = maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalf("Error opening city database: %v", err)
	}
	asnReader, err = maxminddb.Open("GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatalf("Error opening ASN database: %v", err)
	}
	cnReader, err = maxminddb.Open("GeoCN.mmdb")
	if err != nil {
		log.Fatalf("Error opening CN database: %v", err)
	}
}
