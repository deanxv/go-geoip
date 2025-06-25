package common

import (
	"github.com/oschwald/maxminddb-golang"
	"sync"
	"time"
)

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.1.0"            // this hard coding will be replaced automatically when building, no need to manually change

var (
	CityReader, AsnReader, CnReader *maxminddb.Reader
	Mu                              sync.Mutex
)
