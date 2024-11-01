package main

import (
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
	"io"
	"ip2region-geoip/common"
	"ip2region-geoip/common/config"
	logger "ip2region-geoip/common/loggger"
	"ip2region-geoip/middleware"
	"ip2region-geoip/router"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	logger.SetupLogger()
	logger.SysLog(fmt.Sprintf("ip2region-geoip %s started", common.Version))

	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.DebugEnabled {
		logger.SysLog("running in debug mode")
	}
	var err error

	server := gin.New()
	server.Use(gin.Recovery())
	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)
	store := cookie.NewStore([]byte(config.SessionSecret))
	server.Use(sessions.Sessions("session", store))

	router.SetRouter(server)
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	go scheduleDatabaseUpdate()

	err = server.Run(":" + port)
	if err != nil {
		logger.FatalLog("failed to start HTTP server: " + err.Error())
	}
}

//func main() {
//	r := gin.Default()
//
//	// 处理没有参数的情况，使用请求方的 IP
//	r.GET("/ip", func(c *gin.Context) {
//		ip := getRealClientIP(c)
//		info := getIpInfo(ip)
//		c.JSON(http.StatusOK, info)
//	})
//
//	r.GET("/ip/:ip", func(c *gin.Context) {
//		ip := c.Param("ip")
//		if ip == "" {
//			ip = c.ClientIP()
//		}
//		info := getIpInfo(ip)
//		c.JSON(http.StatusOK, info)
//	})
//
//
//	if err := r.Run(":7099"); err != nil {
//		log.Fatal(err)
//	}
//}

func loadDatabases() {
	// 检查文件是否存在，如果不存在则下载
	downloadAndSave("GeoLite2-City.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb")
	downloadAndSave("GeoLite2-ASN.mmdb", "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb")
	downloadAndSave("GeoCN.mmdb", "http://github.com/ljxi/GeoCN/releases/download/Latest/GeoCN.mmdb")

	var err error
	common.Mu.Lock()
	defer common.Mu.Unlock()
	common.CityReader, err = maxminddb.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalf("Error opening city database: %v", err)
	}
	common.AsnReader, err = maxminddb.Open("GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatalf("Error opening ASN database: %v", err)
	}
	common.CnReader, err = maxminddb.Open("GeoCN.mmdb")
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
	// 立即执行一次数据库加载
	loadDatabases()

	// 设置定时任务
	for {
		nextUpdateTime := getNextSundayLastSecond()
		durationUntilUpdate := time.Until(nextUpdateTime)
		log.Printf("Next database update scheduled at %s, which is in %v.", nextUpdateTime, durationUntilUpdate)

		timer := time.NewTimer(durationUntilUpdate)
		<-timer.C

		log.Println("Updating databases...")
		loadDatabases()
	}
}

// getNextSundayLastSecond calculates the next Sunday last second time from now
func getNextSundayLastSecond() time.Time {
	now := time.Now()
	// Calculate how many days to next Sunday (0 = Sunday, 1 = Monday, ..., 6 = Saturday)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 && now.Hour() >= 0 { // If today is Sunday and it's past midnight, wait another week
		daysUntilSunday = 7
	}
	// Set to next Sunday last second
	nextSundayLastSecond := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
	return nextSundayLastSecond
}
