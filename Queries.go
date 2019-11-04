package main

import (
	"strconv"
)

func insertIPs(token, note string, ips []IPset) int {
	//check
	// - user valid
	// - user has ip already inserted
	// - ip already in db ? update : insert

	if len(ips) > 300 {
		return -3
	}

	uid := IsUserValid(token)
	if uid <= 0 {
		return -1
	}

	iplist := concatIPList(ips)

	ip2list := ""
	for _, ip := range ips {
		ip2list += "(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=\"" + ip.IP + "\"),"
	}
	ip2list = ip2list[:len(ip2list)-1]

	sqlGetInsertedIps :=
		"SELECT BlockedIP.ip FROM Reporter " +
			"JOIN BlockedIP on (BlockedIP.pk_id = Reporter.ip) " +
			"WHERE " +
			"BlockedIP.deleted=0 " +
			"AND " +
			"Reporter.ip in (" + ip2list + ") " +
			"AND " +
			"Reporter.reporterID=?"

	sqlGetWhitelisted := "SELECT ip FROM `IPwhitelist` WHERE ip in (" + iplist + ")"
	var alreadyInsertedIps []string

	err := queryRows(&alreadyInsertedIps, sqlGetWhitelisted)
	if err != nil {
		return -2
	}

	err = queryRows(&alreadyInsertedIps, sqlGetInsertedIps, uid)
	if err != nil {
		return -2
	}
	note = EscapeSpecialChars(note)
	if len(note) == 0 {
		note = "NULL"
	}
	ownIP := getOwnIP()
	for _, ip := range ips {
		valid, _ := isIPValid(ip.IP)
		isAlreadyInserted := contains(alreadyInsertedIps, ip.IP)
		if !valid || ip.IP == ownIP || isAlreadyInserted {
			if isAlreadyInserted && ip.Reason > 1 {
				err := execDB(
					"UPDATE Reporter SET reason=? WHERE ip=(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?) AND reporterID=? AND reason<?",
					ip.Reason,
					ip.IP,
					uid,
					ip.Reason,
				)
				if err != nil {
					LogCritical("Update error: " + err.Error())
				}
			}
			if isAlreadyInserted {
				err := execDB(
					"UPDATE Reporter SET note=? WHERE note IS NULL AND ip=(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?)",
					note,
					ip.IP,
				)
				if err != nil {
					LogCritical("Update error: " + err.Error())
				}
			}
			ips = removeIP(ips, ip.IP)
		}
	}

	if len(ips) > 0 {
		valuesBlockedIPs := ""
		for _, ip := range ips {
			valuesBlockedIPs += "(\"" + ip.IP + "\"," + strconv.Itoa(ip.Valid) + "),"
		}
		valuesBlockedIPs = valuesBlockedIPs[:len(valuesBlockedIPs)-1]
		sqlUpdateIps := "INSERT INTO BlockedIP (ip, validated) VALUES " + valuesBlockedIPs + " ON DUPLICATE KEY UPDATE reportCount=reportCount+1, deleted=0"
		err = execDB(sqlUpdateIps)
		if err != nil {
			return -2
		}

		sqlInsertReporter :=
			"INSERT INTO Reporter (Reporter.reporterID, Reporter.ip, reason, note) VALUES "
		note = "\"" + note + "\""
		repData := ""
		for _, ip := range ips {
			repData += "(" + strconv.Itoa(uid) + ",(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=\"" + ip.IP + "\")," + strconv.Itoa(ip.Reason) + ", " + note + "),"
		}
		err = execDB(sqlInsertReporter + repData[:len(repData)-1])
		if err != nil {
			return -2
		}

		sqlUpdateUserReportCount := "UPDATE User SET reportedIPs=reportedIPs+?, lastReport=CURRENT_TIMESTAMP WHERE pk_id=?"
		err = execDB(sqlUpdateUserReportCount, len(ips), uid)
		if err != nil {
			return -2
		}

		doAnalytics(ips)
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

	query :=
		"SELECT ip,deleted " +
			"FROM BlockedIP " +
			"WHERE " +
			"(lastReport >= FROM_UNIXTIME(?) OR firstReport >= FROM_UNIXTIME(?)) "

	if filter.MinReason > 0 {
		query += "AND (SELECT AVG(reason) FROM Reporter WHERE Reporter.ip=BlockedIP.pk_id) >= " + strconv.FormatFloat(filter.MinReason, 'f', 1, 32) + " "
	}

	if filter.MinReports > 0 {
		query += "AND reportCount >= " + strconv.Itoa(filter.MinReports) + " "
	}

	if filter.ProxyAllowed == -1 {
		query += "AND isProxy=0 "
	}

	if filter.Since == 0 {
		query += "AND deleted=0 "
	}

	if filter.OnlyValidatedIPs == -1 {
		query += "AND validated=1 "
	}

	if filter.MaxIPs > 0 {
		query += "LIMIT " + strconv.FormatUint(uint64(filter.MaxIPs), 10)
	}

	var iplist []IPList
	err := queryRows(&iplist, query, filter.Since, filter.Since)

	if err != nil {
		LogCritical("Executing fetch: " + err.Error())
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
