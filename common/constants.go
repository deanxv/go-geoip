package common

import (
	"github.com/oschwald/maxminddb-golang"
	"sync"
	"time"
)

var StartTime = time.Now().Unix() // unit: second
var Version = "v1.0.0"            // this hard coding will be replaced automatically when building, no need to manually change

var (
	CityReader, AsnReader, CnReader *maxminddb.Reader
	AsnMap                          = map[uint]string{
		9812: "东方有线", 9389: "中国长城", 17962: "天威视讯",
		// Add more mappings as needed
	}
	Mu sync.Mutex
)
