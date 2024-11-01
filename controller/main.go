package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"ip2region-geoip/common"
	"ip2region-geoip/model"
	"log"
	"net"
	"net/http"
	"strings"
)

func Ip(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		ip = c.ClientIP()
	}
	info := getIpInfo(ip)
	common.SendResponse(c, http.StatusOK, 0, "success", info)
}

func IpNoArgs(c *gin.Context) {
	ip := getRealClientIP(c)
	info := getIpInfo(ip)
	common.SendResponse(c, http.StatusOK, 0, "success", info)
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

func getAsInfo(number uint) string {
	return common.AsnMap[number]
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

	var asn model.ASN
	common.Mu.Lock()
	defer common.Mu.Unlock()
	if err := common.AsnReader.Lookup(parsedIP, &asn); err == nil {
		asInfo := map[string]interface{}{
			"number": asn.Number,
			"name":   asn.Organization,
		}
		if info := getAsInfo(asn.Number); info != "" {
			asInfo["info"] = info
		}
		ret["as"] = asInfo
	}

	var city model.City
	if network, ok, err := common.CityReader.LookupNetwork(parsedIP, &city); err == nil && ok {
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
		if city.Location.Latitude != 0 && city.Location.Longitude != 0 {
			ret["latitude"] = city.Location.Latitude
			ret["longitude"] = city.Location.Longitude
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

	var geoCN model.GeoCN
	common.Mu.Lock()
	defer common.Mu.Unlock()
	if network, ok, err := common.CnReader.LookupNetwork(parsedIP, &geoCN); err == nil && ok {
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
