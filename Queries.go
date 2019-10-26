package main

import (
	"fmt"
	"strconv"
)

func insertIPs(token string, ips []IPset) int {
	//check
	// - user valid
	// - user has ip already inserted
	// - ip already in db ? update : insert

	sqlCheckUserValid := "SELECT User.pk_id FROM User WHERE token=? AND User.isValid=1"

	var uid int
	err := queryRow(&uid, sqlCheckUserValid, token)
	if err != nil && uid > 0 {
		return -1
	}

	iplist := concatIPList(ips)

	sqlGetInsertedIps := "SELECT BlockedIP.ip FROM `User` " +
		"JOIN Reporter on Reporter.reporterID = User.pk_id " +
		"JOIN BlockedIP on BlockedIP.ip = Reporter.ip " +
		"WHERE User.token = ? AND Reporter.ip in (" + iplist + ")"

	var alreadyInsertedIps []string

	err = queryRows(&alreadyInsertedIps, sqlGetInsertedIps, token)

	if err != nil {
		return -2
	}

	for _, ip := range alreadyInsertedIps {
		ips = removeIP(ips, ip)
	}

	if len(ips) > 0 {
		iplist := concatIPList(ips)
		fmt.Println(iplist)
		valuesBlockedIPs := ""
		for _, ip := range ips {
			valuesBlockedIPs += "(\"" + ip.IP + "\"," + strconv.Itoa(ip.Reason) + "),"
		}
		sqlUpdateIps := "INSERT INTO BlockedIP (ip, reason) VALUES " + valuesBlockedIPs[:len(valuesBlockedIPs)-1] + " ON DUPLICATE KEY UPDATE reportCount=reportCount+1"
		err = execDB(sqlUpdateIps)
		if err != nil {
			return -2
		}
		sqlInsertReporter := "INSERT INTO Reporter (Reporter.reporterID, Reporter.ip) SELECT " + strconv.Itoa(uid) + ",BlockedIP.ip FROM BlockedIP WHERE BlockedIP.ip in (" + iplist + ")"
		err = execDB(sqlInsertReporter)
		if err != nil {
			return -2
		}
	}

	return 1
}
