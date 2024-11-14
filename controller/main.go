package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-geoip/common"
	logger "go-geoip/common/loggger"
	"go-geoip/model"
	"net"
	"net/http"
	"strings"
)

// IP查询
// @Summary IP查询
// @Description IP查询
// @Tags IP查询
// @Produce json
// @Param ip path string true "IP address"
// @Success 200 {object} model.IPInfoResponse "Successful response"
// @Router /ip/{ip} [get]
func Ip(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		ip = c.ClientIP()
	}
	handleIpInfoResponse(c, ip)
}

func IpNoArgs(c *gin.Context) {
	ip := getRealClientIP(c)
	handleIpInfoResponse(c, ip)
}

func handleIpInfoResponse(c *gin.Context, ip string) {
	info, err := getIpInfo(ip)
	if err != nil {
		common.SendResponse(c, http.StatusInternalServerError, 1, "error", err.Error())
		return
	}
	common.SendResponse(c, http.StatusOK, 0, "success", info)
}

func getRealClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			realIP := strings.TrimSpace(ips[0])
			logger.Info(c, fmt.Sprintf("X-Forwarded-For IP: %s", realIP))
			return realIP
		}
	}
	if xrip := c.GetHeader("X-Real-IP"); xrip != "" {
		logger.Info(c, fmt.Sprintf("X-Real-IP: %s", xrip))
		return xrip
	}
	clientIP := c.ClientIP()
	logger.Info(c, fmt.Sprintf("Default ClientIP: %s", clientIP))
	return clientIP
}

func getIpInfo(ip string) (*model.IPInfoResponse, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return nil, fmt.Errorf("invalid IP address")
	}

	info := &model.IPInfoResponse{IP: ip}

	common.Mu.Lock()
	defer common.Mu.Unlock()

	if err := populateASInfo(parsedIP, info); err != nil {
		return nil, err
	}

	if err := populateCityInfo(parsedIP, info); err != nil {
		return nil, err
	}

	if info.Country == "中国" {
		populateCnInfo(ip, info)
	}

	return info, nil
}

func populateASInfo(parsedIP net.IP, info *model.IPInfoResponse) error {
	var asn model.ASN
	if err := common.AsnReader.Lookup(parsedIP, &asn); err != nil {
		return err
	}
	info.AS = asn.Organization
	return nil
}

func populateCityInfo(parsedIP net.IP, info *model.IPInfoResponse) error {
	var city model.City
	if network, ok, err := common.CityReader.LookupNetwork(parsedIP, &city); err == nil && ok {
		info.Addr = network.String()
		info.Country = getCountry(city.Country.Names)
		info.RegisteredCountry = getCountry(city.RegisteredCountry.Names)
		info.Latitude = city.Location.Latitude
		info.Longitude = city.Location.Longitude
		info.Subdivisions = getSubdivisions(city.Subdivisions)
		info.City = getCityName(city.City.Names)
	} else if err != nil {
		return err
	}
	return nil
}

func populateCnInfo(ip string, info *model.IPInfoResponse) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return
	}

	var geoCN model.GeoCN
	if network, ok, err := common.CnReader.LookupNetwork(parsedIP, &geoCN); err == nil && ok {
		info.Addr = network.String()
		if strings.HasSuffix(geoCN.Province, "市") {
			info.Province = geoCN.Province
			info.City = geoCN.Province
			info.District = geoCN.City
		} else {
			info.Province = geoCN.Province
			info.City = geoCN.City
			info.District = geoCN.Districts
		}

		info.AS = geoCN.ISP
		if geoCN.Net != "" {
			info.AS += " (" + geoCN.Net + ")"
		}
	}
}

func getSubdivisions(subdivisions []model.Subdivision) []string {
	var names []string
	for _, subdivision := range subdivisions {
		if name, exists := subdivision.Names["zh-CN"]; exists {
			names = append(names, name)
		} else {
			names = append(names, subdivision.Names["en"])
		}
	}
	return names
}

func getCityName(names map[string]string) string {
	if name, exists := names["zh-CN"]; exists {
		return name
	}
	return names["en"]
}

func getCountry(names map[string]string) string {
	if name, ok := names["zh-CN"]; ok {
		switch name {
		case "香港", "澳门", "台湾":
			return "中国" + name
		default:
			return name
		}
	}
	return names["en"]
}
