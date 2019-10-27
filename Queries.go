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

	uid := IsUserValid(token)
	if uid <= 0 {
		return -1
	}

	iplist := concatIPList(ips)

	sqlGetInsertedIps := "SELECT BlockedIP.ip FROM `User` " +
		"JOIN Reporter on Reporter.reporterID = User.pk_id " +
		"JOIN BlockedIP on BlockedIP.ip = Reporter.ip " +
		"WHERE User.token = ? AND Reporter.ip in (" + iplist + ")"

	var alreadyInsertedIps []string

	err := queryRows(&alreadyInsertedIps, sqlGetInsertedIps, token)

	if err != nil {
		return -2
	}

	ownIP := getOwnIP()
	for _, ip := range ips {
		valid, _ := isIPValid(ip.IP)
		if !valid || ip.IP == ownIP || contains(alreadyInsertedIps, ip.IP) {
			ips = removeIP(ips, ip.IP)
		}
	}

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

func fetchIPsFromDB(token string, filter FetchFilter) ([]IPList, int) {
	//check
	// - user valid

	uid := IsUserValid(token)
	if uid <= 0 {
		return nil, -1
	}

	//sqlAdditions := ""

	query :=
		"SELECT ip " +
			"FROM BlockedIP " +
			"WHERE " +
			"lastReport > FROM_UNIXTIME(?) "

	if filter.MinReason > 0 {
		query += "AND (SELECT AVG(reason) FROM IPreason WHERE IPreason.ip=BlockedIP.pk_id) >= " + strconv.FormatFloat(filter.MinReason, 'f', 1, 32) + " "
	}

	if filter.MinReports > 0 {
		query += "AND reportCount >= " + strconv.Itoa(filter.MinReports) + " "
	}

	if filter.ProxyAllowed == 0 {
		query += "AND isProxy=0 "
	}

	if filter.MaxIPs > 0 {
		query += "LIMIT " + strconv.FormatUint(uint64(filter.MaxIPs), 10)
	}

	var iplist []IPList
	err := queryRows(&iplist, query, filter.Since)

	if err != nil {
		fmt.Println(err.Error())
		return nil, 1
	}

	return iplist, 0
}

//IsUserValid returns userid if valid or -1 if invalid
func IsUserValid(token string) int {
	sqlCheckUserValid := "SELECT User.pk_id FROM User WHERE token=? AND User.isValid=1"
	var uid int
	err := queryRow(&uid, sqlCheckUserValid, token)
	if err != nil && uid > 0 {
		return -1
	}
	return uid
}
