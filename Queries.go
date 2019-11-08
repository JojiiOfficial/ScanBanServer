package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

//ReportIPcount ip and count
type ReportIPcount struct {
	Count int `db:"c"`
	ID    int `db:"iid"`
}

const batchSize = 30

func insertIPs2(token string, ipdatas []IPData, starttime int64) int {
	uid := IsUserValid(token)
	if uid <= 0 {
		return -1
	}

	for _, ipdata := range ipdatas {
		ipID, reportID, err := insertIP(ipdata, uid)
		_ = ipID
		if err != nil {
			LogCritical("Error inserting ip: " + err.Error())
			continue
		}
		if reportID == -1 {
			reportID, err = insertReport(ipdata, uid)
			if err != nil {
				LogCritical("Error inserting report: " + err.Error())
				continue
			}
		}
		err = execDB("UPDATE Report SET lastReport=CURRENT_TIMESTAMP WHERE pk_id=?", reportID)
		if err != nil {
			LogCritical("Error updating last report: " + err.Error())
			continue
		}
		//IP And report is inserted

		for _, iPPort := range ipdata.Ports {
			if len(iPPort.Times) == 0 || iPPort.Port < 1 || iPPort.Port > 65535 {
				LogInfo("IP data invalid: " + ipdata.IP + ":" + strconv.Itoa(iPPort.Port))
				continue
			}
			batches := make(map[int][]int)
			for _, time := range iPPort.Times {
				pos := (int)(time / batchSize)
				_, ok := batches[pos]
				if !ok {
					batches[pos] = []int{time}
				} else {
					batches[pos] = append(batches[pos], time)
				}
			}
			insertBatch(batches, reportID, iPPort.Port, starttime)
		}
	}

	return 1
}

func insertBatch(batch map[int][]int, reportID, port int, startTime int64) {
	fmt.Println(port, batch)
	for _, b := range batch {
		scanCount := len(b)
		if scanCount == 0 {
			continue
		}
		err := execDB("INSERT INTO ReportPorts (reportID, port, count, scanDate) VALUES(?,?,?,FROM_UNIXTIME(?))",
			reportID, port, scanCount, startTime+int64(min(b)))
		if err != nil {
			LogCritical("Couldn't insert ReportPort: " + err.Error())
		}
	}
}

func min(intsl []int) int {
	if len(intsl) == 0 {
		return 0
	}
	if len(intsl) == 1 {
		return intsl[0]
	}
	ix := intsl[0]
	for _, i := range intsl {
		if i < ix {
			ix = i
		}
	}
	return ix
}

func max(intsl []int) int {
	if len(intsl) == 0 {
		return 0
	}
	if len(intsl) == 1 {
		return intsl[0]
	}
	ix := intsl[0]
	for _, i := range intsl {
		if i > ix {
			ix = i
		}
	}
	return ix
}

func insertReport(ipdata IPData, uid int) (int, error) {
	err := execDB("INSERT INTO Report (ip, reporterID) VALUES((SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?),?)", ipdata.IP, uid)
	if err != nil {
		LogCritical("Couldn't execute insert ip into report: " + err.Error())
		return -1, err
	}
	var id int
	err = queryRow(&id, "SELECT Report.pk_id FROM Report WHERE ip=(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?) AND reporterID=?", ipdata.IP, uid)
	if err != nil {
		return -1, err
	} else if id == 0 {
		return -1, errors.New("report not found")
	}
	return id, nil
}

func insertIP(ipdata IPData, uid int) (IPid int, reportID int, err error) {
	IPid, reportID = -1, -1
	err = nil

	var c ReportIPcount
	err = queryRow(&c, "SELECT COUNT(*) as c, ifnull(pk_id, -1)as iid FROM Report WHERE reporterID=? AND ip=ifnull((SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?),\"\")", uid, ipdata.IP)
	if err != nil {
		return
	}
	if c.Count != 0 {
		reportID = c.ID
		return
	}
	err = execDB("INSERT INTO BlockedIP (ip, validated) VALUES (?,0) ON DUPLICATE KEY UPDATE reportCount=reportCount+1, deleted=0", ipdata.IP)
	if err != nil {
		return
	}
	doAnalytics(ipdata)
	err = queryRow(&IPid, "SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?", ipdata.IP)
	if err != nil {
		return
	}
	reportID = c.ID
	return
}

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

	note = EscapeSpecialChars(strings.Trim(note, " "))

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
			if isAlreadyInserted && len(note) > 0 {
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
			LogCritical("Couldn't insert into BlockedIP: " + err.Error())
			return -2
		}

		sqlInsertReporter :=
			"INSERT INTO Reporter (Reporter.reporterID, Reporter.ip, reason, note) VALUES "

		if len(note) == 0 {
			note = "NULL"
		} else {
			note = "\"" + note + "\""
		}
		repData := ""
		for _, ip := range ips {
			repData += "(" + strconv.Itoa(uid) + ",(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=\"" + ip.IP + "\")," + strconv.Itoa(ip.Reason) + ", " + note + "),"
		}
		err = execDB(sqlInsertReporter + repData[:len(repData)-1])
		if err != nil {
			LogCritical("Couldn't insert into Reporter: " + err.Error())
			return -2
		}

		sqlUpdateUserReportCount := "UPDATE User SET reportedIPs=reportedIPs+?, lastReport=CURRENT_TIMESTAMP WHERE pk_id=?"
		err = execDB(sqlUpdateUserReportCount, len(ips), uid)
		if err != nil {
			LogCritical("Couldn't update user: " + err.Error())
			return -2
		}
		LogInfo("Added " + strconv.Itoa(len(ips)) + " new IPs with note " + note + " from " + strconv.Itoa(uid))
		//doAnalytics(ips)
	} else {
		LogInfo("Reported but no new IP added")
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
	} else if err != nil {
		panic(err)
	}
	return uid
}
