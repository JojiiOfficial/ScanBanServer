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

	for _, ip := range ips {
		valid, _ := isIPValid(ip.IP)
		if !valid || ip.IP == externIP || contains(alreadyInsertedIps, ip.IP) {
			ips = removeIP(ips, ip.IP)
		}
	}

	// for _, ip := range alreadyInsertedIps {
	// 	ips = removeIP(ips, ip)
	// }

	if len(ips) > 0 {
		iplist := concatIPList(ips)

		valuesBlockedIPs := ""
		for _, ip := range ips {
			valuesBlockedIPs += "(\"" + ip.IP + "\"),"
		}
		valuesBlockedIPs = valuesBlockedIPs[:len(valuesBlockedIPs)-1]
		sqlUpdateIps := "INSERT INTO BlockedIP (ip) VALUES " + valuesBlockedIPs + " ON DUPLICATE KEY UPDATE reportCount=reportCount+1"
		err = execDB(sqlUpdateIps)
		if err != nil {
			return -2
		}

		sqlInsertReporter := "INSERT INTO Reporter (Reporter.reporterID, Reporter.ip) SELECT " + strconv.Itoa(uid) + ",BlockedIP.ip FROM BlockedIP WHERE BlockedIP.ip in (" + iplist + ")"
		err = execDB(sqlInsertReporter)
		if err != nil {
			return -2
		}

		sqlSelectIPs := "SELECT ip, pk_id FROM BlockedIP WHERE ip in (" + iplist + ")"
		var ipids []IPID
		err = queryRows(&ipids, sqlSelectIPs)
		if err != nil {
			return -2
		}
		IPreasonData := ""
		for _, ip := range ips {
			IPreasonData += "(" + strconv.Itoa(ipidFromIP(ipids, ip.IP).ID) + ", " + strconv.Itoa(ip.Reason) + "," + strconv.Itoa(uid) + "),"
		}
		IPreasonData = IPreasonData[:len(IPreasonData)-1]
		fmt.Println(IPreasonData)
		err = execDB("INSERT INTO IPreason (ip, reason, author) VALUES " + IPreasonData)
		if err != nil {
			return -2
		}

		sqlUpdateUserReportCount := "UPDATE User SET reportedIPs=reportedIPs+?, lastReport=CURRENT_TIMESTAMP WHERE pk_id=?"
		err = execDB(sqlUpdateUserReportCount, len(ips), uid)
		if err != nil {
			return -2
		}
	}

	return 1
}
