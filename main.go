package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/getlantern/geo"
)

const (
	geolite2_url   = "https://d254wvfcgkka1d.cloudfront.net/app/geoip_download?license_key=%s&edition_id=GeoLite2-Country&suffix=tar.gz"
	geoip2_isp_url = "https://d254wvfcgkka1d.cloudfront.net/app/geoip_download?license_key=%s&edition_id=GeoIP2-ISP&suffix=tar.gz"
)

var (
	countryLookup geo.CountryLookup = &geo.NoLookup{}
	asnLookup     geo.ISPLookup     = &geo.NoLookup{}
)

func dumpIP(ip net.IP) {
	cc := countryLookup.CountryCode(ip)
	asn := asnLookup.ASN(ip)
	fmt.Printf("%v %v %v\n", ip, cc, asn)
}

func dumpCIDR(cidrIP net.IP, ipnet *net.IPNet) {
	for ip := cidrIP.Mask(ipnet.Mask); ipnet.Contains(ip); incIP(ip) {
		if ip.IsGlobalUnicast() {
			dumpIP(ip)
		}
	}
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func main() {
	licenseKey := os.Getenv("MAXMIND_LICENSE_KEY")
	cl := geo.FromWeb(fmt.Sprintf(geolite2_url, licenseKey), "GeoLite2-Country.mmdb", 48*time.Hour, "GeoLite2-Country.mmdb", geo.CountryCode)
	al := geo.FromWeb(fmt.Sprintf(geoip2_isp_url, licenseKey), "GeoIP2-ISP.mmdb", 48*time.Hour, "GeoIP2-ISP.mmdb", geo.ASN)
	<-cl.Ready()
	<-al.Ready()
	countryLookup = cl
	asnLookup = al

	for _, arg := range os.Args[1:] {
		ip, network, err := net.ParseCIDR(arg)
		if err == nil {
			dumpCIDR(ip, network)
		} else {
			ip := net.ParseIP(arg)
			if ip == nil {
				fmt.Errorf("failed to parse argument '%s' as an IP or CIDR block...", arg)
				return
			}
			dumpIP(ip)
		}
	}
}
