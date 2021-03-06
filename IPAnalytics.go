package main

import (
	"net"
	"strings"
	"time"

	"github.com/theckman/go-ipdata"
)

func hostnameCheck(c chan *string, ip string) {
	for i := 0; i <= 1; i++ {
		LogInfo("Lookup hostname try")
		addr, err := net.LookupAddr(ip)
		if err == nil && len(addr) > 0 {
			c <- &addr[0]
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	empty := ""
	c <- &empty
}

func ipDataCheck(c chan *ipdata.IP, ip string) {
	data, err := ipdataClient.Lookup(ip)
	if err != nil {
		LogCritical("Error looking up ip: " + ip + "  " + err.Error())
		ipdataClient = nil
		c <- nil
		return
	}
	c <- &data
}

func doAnalyticsLecacy(ips []IPset) {
	go (func(ips []IPset) {
		connectIPDataClient(config)
		for _, ip := range ips {
			runAnalytic(ip.IP)
		}
	})(ips)
}

func doAnalytics(ip IPData) {
	connectIPDataClient(config)
	go (func(ip IPData) {
		runAnalytic(ip.IP)
	})(ip)
}

func runAnalytic(ip string) {
	cHost := make(chan *string)
	go hostnameCheck(cHost, ip)

	if ipdataClient != nil {
		cInf := make(chan *ipdata.IP)
		go ipDataCheck(cInf, ip)
		hostname, ipdata := <-cHost, <-cInf
		if ipdata == nil {
			updateWithHostname(hostname, ip)
		} else {
			updateWithIPdata(hostname, *ipdata, ip)
		}
	} else {
		updateWithHostname(<-cHost, ip)
	}
}

func getValidHostnameKeys() []string {
	var list []string
	err := queryRows(&list, "SELECT keyword FROM KnownHostname")
	if err != nil {
		LogCritical("Couldn't load Keywords: " + err.Error())
		return []string{}
	}
	return list
}

func validateHostname(hostname string, validateList []string) bool {
	for _, key := range validateList {
		prefixStar := strings.HasPrefix(key, "*")
		suffixStar := strings.HasSuffix(key, "*")
		key = strings.ReplaceAll(key, "*", "")
		if prefixStar && suffixStar {
			if strings.Contains(hostname, key) {
				return true
			}
		}

		if !prefixStar && suffixStar {
			if strings.HasPrefix(hostname, key) {
				return true
			}
		}

		if prefixStar && !suffixStar {
			if strings.HasSuffix(hostname, key) {
				return true
			}
		}
		if !prefixStar && !suffixStar {
			if hostname == key {
				return true
			}
		}
	}
	return false
}

func updateWithIPdata(hostname *string, ipdata ipdata.IP, ip string) {
	if len(*hostname) == 0 {
		hostname = nil
	} else {
		updateValide(*hostname, ip)
	}
	isProxy := 0
	if ipdata.Threat.IsAnonymous {
		isProxy = 1
	}
	var domain *string
	domain = nil
	if len(ipdata.ASN.Domain) > 0 {
		domain = &ipdata.ASN.Domain
	}
	knownAbuser := "0"
	if ipdata.Threat.IsKnownAbuser {
		knownAbuser = "1"
	}
	knownHacker := "0"
	if ipdata.Threat.IsKnownAttacker {
		knownHacker = "1"
	}
	err := execDB(
		"UPDATE BlockedIP set Hostname=?, isProxy=?, type=(SELECT pk_id FROM IPtype WHERE type=?), domain=?, knownAbuser=?,	knownHacker=? WHERE ip=?",
		hostname,
		isProxy,
		ipdata.ASN.Type,
		domain,
		knownAbuser,
		knownHacker,
		ip,
	)

	if err != nil {
		LogCritical("Error updating host and ipdata")
	}
}

func updateWithHostname(hostname *string, ip string) {
	if len(*hostname) == 0 {
		hostname = nil
	} else {
		updateValide(*hostname, ip)
	}
	err := execDB("UPDATE BlockedIP SET Hostname=? WHERE ip=?", hostname, ip)
	if err != nil {
		LogCritical("Error updating hostname!")
	}
}

func updateValide(hostname, ip string) {
	allowedKeys := getValidHostnameKeys()
	if validateHostname(hostname, allowedKeys) {
		err := execDB("UPDATE BlockedIP SET validated=1 WHERE ip=?", ip)
		if err != nil {
			LogCritical("Couldn't update valid=1 on host: " + err.Error())
		}
	}
}
