package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/oschwald/maxminddb-golang"
	"go-geoip/common"
	"go-geoip/common/config"
	logger "go-geoip/common/loggger"
	"go-geoip/middleware"
	"go-geoip/router"
)

const (
	cityDBDefaultURL = "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-City.mmdb"
	asnDBURL         = "https://github.com/P3TERX/GeoLite.mmdb/raw/download/GeoLite2-ASN.mmdb"
	cnDBURL          = "http://github.com/ljxi/GeoCN/releases/download/Latest/GeoCN.mmdb"
	sessionName      = "session"
)

func main() {
	logger.SetupLogger()
	logger.SysLog(fmt.Sprintf("go-geoip %s started", common.Version))

	setGinMode()
	logDebugMode()

	server := setupServer()

	go scheduleDatabaseUpdate()

	runServer(server)
}

func setGinMode() {
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func logDebugMode() {
	if config.DebugEnabled {
		logger.SysLog("running in debug mode")
	}
}

func setupServer() *gin.Engine {
	server := gin.New()
	server.Use(gin.Recovery(), middleware.RequestId())
	middleware.SetUpLogger(server)

	store := cookie.NewStore([]byte(config.SessionSecret))
	server.Use(sessions.Sessions(sessionName, store))

	router.SetRouter(server)
	return server
}

func runServer(server *gin.Engine) {
	port := getPort()
	if err := server.Run(":" + port); err != nil {
		logger.FatalLog("failed to start HTTP server: %v", err)
	}
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return strconv.Itoa(*common.Port)
}

func loadDatabases() {
	downloadAndSave("GeoIP-City.mmdb", getCityDBURL())
	downloadAndSave("Geo-ASN.mmdb", asnDBURL)
	downloadAndSave("GeoCN.mmdb", cnDBURL)

	openDatabases()
}

func getCityDBURL() string {
	if config.CityDBRemoteUrl != "" {
		return config.CityDBRemoteUrl
	}
	return cityDBDefaultURL
}

func downloadAndSave(filename, url string) {
	logger.SysLog(fmt.Sprintf("Downloading %s...", filename))
	resp, err := http.Get(url)
	if err != nil {
		logger.FatalLog("Failed to download %s: %v", filename, err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		logger.FatalLog("Failed to create file %s: %v", filename, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		logger.FatalLog("Failed to save file %s: %v", filename, err)
	}
	logger.SysLog(fmt.Sprintf("Downloaded and saved %s successfully", filename))
}

func openDatabases() {
	var err error
	common.Mu.Lock()
	defer common.Mu.Unlock()

	common.CityReader, err = maxminddb.Open("GeoIP-City.mmdb")
	if err != nil {
		logger.FatalLog("Error opening city database: %v", err)
	}

	common.AsnReader, err = maxminddb.Open("Geo-ASN.mmdb")
	if err != nil {
		logger.FatalLog("Error opening ASN database: %v", err)
	}

	common.CnReader, err = maxminddb.Open("GeoCN.mmdb")
	if err != nil {
		logger.FatalLog("Error opening CN database: %v", err)
	}
}

func scheduleDatabaseUpdate() {
	loadDatabases()

	for {
		nextUpdateTime := getNextSundayLastSecond()
		durationUntilUpdate := time.Until(nextUpdateTime)
		logger.SysLog(fmt.Sprintf("Next database update scheduled at %s, which is in %v.", nextUpdateTime, durationUntilUpdate))

		timer := time.NewTimer(durationUntilUpdate)
		<-timer.C

		logger.SysLog("Updating databases...")
		loadDatabases()
	}
}

func getNextSundayLastSecond() time.Time {
	now := time.Now()
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 && now.Hour() >= 0 {
		daysUntilSunday = 7
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Add(time.Duration(daysUntilSunday) * 24 * time.Hour)
}
