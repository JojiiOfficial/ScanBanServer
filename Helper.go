package main

import (
	"html"
	"net"
	"net/http"
	"strings"
)

var reservedIPs = []string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.88.99.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"224.0.0.0/4",
	"240.0.0.0/4",
}

//returns if ip is valid and a reason
func isIPValid(ip string) (bool, int) {
	pip := net.ParseIP(ip)
	if pip.To4() == nil {
		return false, 0
	}
	for _, reservedIP := range reservedIPs {
		_, subnet, err := net.ParseCIDR(reservedIP)
		if err != nil {
			panic(err)
		}
		if subnet.Contains(pip) {
			return false, -1
		}
	}
	return true, 1
}

func removeIP(iplist []IPset, ip string) []IPset {
	var newIPs []IPset
	for _, cip := range iplist {
		if cip.IP != ip {
			newIPs = append(newIPs, cip)
		}
	}
	return newIPs
}

func concatIPList(ips []IPset) string {
	iplist := "\""
	for _, ip := range ips {
		iplist += ip.IP + "\",\""
	}
	return iplist[:len(iplist)-2]
}

func contains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

//EscapeSpecialChars avoid sqlInjection
func EscapeSpecialChars(inp string) string {
	if len(inp) == 0 {
		return ""
	}
	toReplace := []string{"'", "`", "\""}
	for _, i := range toReplace {
		inp = strings.ReplaceAll(inp, i, "")
	}
	return html.EscapeString(inp)
}

func getIPFromHTTPrequest(r *http.Request) string {
	ipheader := []string{"X-Forwarded-For", "X-Real-Ip", "HTTP_CLIENT_IP", "HTTP_X_FORWARDED_FOR", "HTTP_X_FORWARDED", "HTTP_X_CLUSTER_CLIENT_IP", "HTTP_FORWARDED_FOR", "HTTP_FORWARDED", "REMOTE_ADDR"}
	var repIP string
	for _, header := range ipheader {
		cip := r.Header.Get(header)
		cip = strings.Trim(cip, " ")
		if len(cip) > 0 {
			repIP = cip
			break
		}
	}
	if len(strings.Trim(repIP, " ")) == 0 {
		repIP = r.RemoteAddr
	}
	if strings.Contains(repIP, ":") {
		repIP = repIP[:(strings.LastIndex(repIP, ":"))]
	}
	return repIP
}

func isInStringArray(str string, arr []string, trim bool) bool {
	if trim {
		str = strings.Trim(str, " ")
	}
	for _, s := range arr {
		z := s
		if trim {
			z = strings.Trim(z, " ")
		}
		if z == str {
			return true
		}
	}
	return false
}
