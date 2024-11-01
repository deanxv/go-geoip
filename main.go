package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
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
	mu sync.Mutex
)

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
	mu.Lock()
	defer mu.Unlock()
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
	mu.Lock()
	defer mu.Unlock()
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

	// 处理没有参数的情况，使用请求方的 IP
	r.GET("/ip", func(c *gin.Context) {
		ip := getRealClientIP(c)
		info := getIpInfo(ip)
		c.JSON(http.StatusOK, info)
	})

	r.GET("/ip/:ip", func(c *gin.Context) {
		ip := c.Param("ip")
		if ip == "" {
			ip = c.ClientIP()
		}
		info := getIpInfo(ip)
		c.JSON(http.StatusOK, info)
	})

	go scheduleDatabaseUpdate()

	if err := r.Run(":7099"); err != nil {
		log.Fatal(err)
	}
}

// getRealClientIP 尝试从多个头部信息获取真实 IP，如果没有则使用 c.ClientIP()
func getRealClientIP(c *gin.Context) string {
	// 尝试从 X-Forwarded-For 头获取
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// 分割 X-Forwarded-For 字符串，获取第一个 IP 地址
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			realIP := strings.TrimSpace(ips[0]) // 清除可能的空格
			log.Printf("X-Forwarded-For IP: %s", realIP)
			return realIP
		}
	}
	// 尝试从 X-Real-IP 头获取
	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		log.Printf("X-Real-IP: %s", xrip)
		return xrip
	}
	// 默认使用 Gin 的 ClientIP 方法
	clientIP := c.ClientIP()
	log.Printf("Default ClientIP: %s", clientIP)
	return clientIP
}

func init() {
	loadDatabases()
}

func loadDatabases() {
	// 检查文件是否存在，如果不存在则下载
	downloadAndSave("GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb")
	downloadAndSave("GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb")
	downloadAndSave("GeoCN.mmdb", "http://github.com/ljxi/GeoCN/releases/download/Latest/GeoCN.mmdb")

	var err error
	mu.Lock()
	defer mu.Unlock()
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

func downloadAndSave(filename, url string) {
	log.Printf("Downloading %s...", filename)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to download %s: %v", filename, err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file %s: %v", filename, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Fatalf("Failed to save file %s: %v", filename, err)
	}
	log.Printf("Downloaded and saved %s successfully", filename)
}

func scheduleDatabaseUpdate() {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Updating databases...")
			loadDatabases()
		}
	}
}
