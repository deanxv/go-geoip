package config

import (
	"go-geoip/common/env"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

var DebugEnabled = strings.ToLower(os.Getenv("DEBUG")) == "true"

var SessionSecret = uuid.New().String()

var RateLimitKeyExpirationDuration = 20 * time.Minute
var SwaggerEnable = os.Getenv("SWAGGER_ENABLE")
var ApiSecret = os.Getenv("API_SECRET")
var ApiSecrets = strings.Split(os.Getenv("API_SECRET"), ",")

var CityDBRemoteUrl = os.Getenv("CITY_DB_REMOTE_URL")

var (
	RequestRateLimitNum            = env.Int("REQUEST_RATE_LIMIT", 120)
	RequestRateLimitDuration int64 = 1 * 60
)
