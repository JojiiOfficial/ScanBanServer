package main

import (
	"fmt"
	"net"
	"time"

	"github.com/theckman/go-ipdata"
)

func hostnameCheck(c chan string, ip IPset) {
	for i := 0; i <= 1; i++ {
		fmt.Println("Lookup hostname try", i)
		addr, err := net.LookupAddr(ip.IP)
		if err == nil && len(addr) > 0 {
			c <- addr[0]
			return
		}
		time.Sleep(500 * time.Millisecond)
	}
	c <- ""
}

func ipDataCheck(c chan *ipdata.IP, ip IPset) {
	data, err := ipdataClient.Lookup(ip.IP)
	if err != nil {
		fmt.Println("Error looking up ip:", ip.IP, "", err.Error())
		ipdataClient = nil
		c <- nil
		return
	}
	c <- &data
}

func doAnalytics(ips []IPset) {
	go (func(ips []IPset) {
		for _, ip := range ips {
			cHost := make(chan string)
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
	})(ips)
}

func updateWithIPdata(hostname string, ipdata ipdata.IP, ip IPset) {
	if len(hostname) == 0 {
		hostname = "NULL"
	}
	isProxy := 0
	if ipdata.Threat.IsAnonymous {
		isProxy = 1
	}
	domain := "NULL"
	if len(ipdata.ASN.Domain) > 0 {
		domain = ipdata.ASN.Domain
	}
	err := execDB(
		"UPDATE BlockedIP set Hostname=?, isProxy=?, type=(SELECT pk_id FROM IPtype WHERE type=?), domain=? WHERE ip=?",
		hostname,
		isProxy,
		ipdata.ASN.Type,
		domain,
		ip.IP,
	)
	if err != nil {
		fmt.Println("Error updating host and ipdata")
	}
}

func updateWithHostname(hostname string, ip IPset) {
	err := execDB("UPDATE BlockedIP SET Hostname=? WHERE ip=?", hostname, ip.IP)
	if err != nil {
		fmt.Println("Error updating hostname!")
	}
}
