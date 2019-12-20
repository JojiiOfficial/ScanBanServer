package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

//IPInfo info for Ip
type IPInfo struct {
	Hostname string `json:"Hostname"`
	Type     int    `json:"Type"`
	Asn      string `json:"Asn"`
	Country  string `json:"Country"`
	IsProxy  bool   `json:"IsProxy"`
}

//IPDataRequest request for ip
func IPDataRequest(ip string, ignoreCert bool) *IPInfo {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCert},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("http://codefacility.de:2666/info/get/" + ip)
	if err != nil {
		return nil
	}

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var info IPInfo
	err = json.Unmarshal(d, &info)
	if err != nil {
		fmt.Println("Can't parse json:" + err.Error())
		return nil
	}
	return &info
}
