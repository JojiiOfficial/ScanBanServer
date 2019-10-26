package main

import "fmt"

func insertIPs(token string, ips []IPset) int {
	//check
	// - user valid
	// - user has ip already inserted
	// - ip already in db ? update : insert

	sqlCheckUserValid := "SELECT User.isValid FROM User WHERE token=?"
	var i int
	err := queryRow(&i, sqlCheckUserValid, token)
	fmt.Println(i)
	if err != nil || i != 1 {
		return -1
	}

	iplist := []string{}
	for _, ip := range ips {
		iplist = append(iplist, ip.IP)
	}

	sqlGetInsertedIps := "SELECT BlockedIP.ip FROM `User`" +
		"JOIN Reporter on Reporter.reporterID = User.pk_id" +
		"JOIN BlockedIP on BlockedIP.ip = Reporter.ip" +
		"WHERE User.token = ? AND Reporter.ip in (?)"

	var insertedIps []string

	err = queryRows(&insertedIps, sqlGetInsertedIps, token, iplist)

	if err != nil {
		panic(err)
	}

	for _, ip := range insertedIps {
		fmt.Println(ip)
	}

	return 1
}
