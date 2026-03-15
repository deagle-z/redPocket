package utils

import (
	"log"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

var (
	geoOnce sync.Once
	geoDB   *geoip2.Reader
)

func getGeoDB() *geoip2.Reader {
	geoOnce.Do(func() {
		db, err := geoip2.Open("GeoLite2-Country.mmdb")
		if err != nil {
			log.Printf("[GeoIP] failed to open GeoLite2-Country.mmdb: %v", err)
			return
		}
		geoDB = db
	})
	return geoDB
}

// GetCountryCodeByIP 根据 IP 返回 ISO 国家码（如 IN / BR），失败返回空字符串
func GetCountryCodeByIP(ip string) string {
	db := getGeoDB()
	if db == nil {
		return ""
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	record, err := db.Country(parsed)
	if err != nil {
		return ""
	}
	return record.Country.IsoCode
}
