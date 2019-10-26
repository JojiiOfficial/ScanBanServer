package main

import "strconv"

func insertIPs(token string, ips []IPset) int {
	//check
	// - user valid
	// - user has ip already inserted
	// - ip already in db ? update : insert

	sqlCheckUserValid := "SELECT User.isValid FROM User WHERE token=?"
	var i int
	err := queryRow(&i, sqlCheckUserValid, token)
	if err != nil || i != 1 {
		return -1
	}

	iplist := "\""
	for _, ip := range ips {
		iplist += ip.IP + "\",\""
	}
	iplist = iplist[:len(iplist)-2]

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
		valuesBlockedIPs := ""
		for _, ip := range ips {
			valuesBlockedIPs += "(\"" + ip.IP + "\"," + strconv.Itoa(ip.Reason) + "),"
		}
		sqlUpdateIps := "INSERT INTO BlockedIP (ip, reason) VALUES " + valuesBlockedIPs[:len(valuesBlockedIPs)-1] + " ON DUPLICATE KEY UPDATE reportCount=reportCount+1"

		err = execDB(sqlUpdateIps)
	}

	return 1
}
