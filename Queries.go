package main

import (
	"errors"
	"strconv"
)

//ReportIPcount ip and count
type ReportIPcount struct {
	Count int `db:"c"`
	ID    int `db:"iid"`
}

const batchSize = 30

func insertIPs(token string, ipdatas []IPData, starttime int64) int {
	uid := IsUserValid(token)
	if uid <= 0 {
		return -1
	}

	sqlUpdateUserReportCount := "UPDATE User SET reportedIPs=reportedIPs+?, lastReport=CURRENT_TIMESTAMP WHERE pk_id=?"
	err := execDB(sqlUpdateUserReportCount, len(ipdatas), uid)
	if err != nil {
		LogCritical("Error updating lastReport")
		return -2
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
	values := ""
	for _, b := range batch {
		scanCount := len(b)
		if scanCount == 0 {
			continue
		}
		var rpID int
		err := queryRow(&rpID, "SELECT IFNULL(MAX(pk_id),-1) FROM ReportPorts WHERE scanDate >= ? AND reportID=? AND port=?", startTime-int64(batchSize), reportID, port)
		if err != nil {
			LogCritical("Couldn't get reportPorts in current batch: " + err.Error())
			continue
		}

		if rpID > 0 {
			err = execDB("UPDATE ReportPorts SET count=count+? WHERE pk_id=?", scanCount, rpID)
		} else {
			values += "(" + strconv.Itoa(reportID) + "," + strconv.Itoa(port) + "," + strconv.Itoa(scanCount) + "," + strconv.FormatInt(startTime, 10) + "),"
		}
	}

	if len(values) > 2 {
		err := execDB("INSERT INTO ReportPorts (reportID, port, count, scanDate) VALUES" + values[:len(values)-1])
		if err != nil {
			LogCritical("Couldn't insert ReportPort: " + err.Error())
		}
	}
}

func getIPInfo(ips []string, token string) (int, *[]IPInfoData) {
	uid := IsUserValid(token)
	if uid <= 0 {
		return -1, nil
	}
	ipdata := []IPInfoData{}
	for _, ip := range ips {
		var info []ReportData
		err := queryRows(&info, "SELECT Report.reporterID, User.username, ReportPorts.scanDate, ReportPorts.port, ReportPorts.count FROM `Report`"+
			"JOIN BlockedIP on (BlockedIP.pk_id = Report.ip)"+
			"JOIN User on (User.pk_id = Report.reporterID)"+
			"JOIN ReportPorts on (ReportPorts.reportID = Report.pk_id)"+
			"WHERE BlockedIP.ip=? ORDER BY ReportPorts.scanDate ASC", ip)
		if err != nil {
			LogCritical("Error getting info: " + err.Error())
			return 2, nil
		}
		ipdata = append(ipdata, IPInfoData{
			IP:      ip,
			Reports: info,
		})
	}

	return 1, &ipdata
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
	var ce int
	err = queryRow(&ce, "SELECT COUNT(*) FROM BlockedIP WHERE ip=?", ipdata.IP)
	if err != nil {
		return
	}
	err = execDB("INSERT INTO BlockedIP (ip, validated) VALUES (?,0) ON DUPLICATE KEY UPDATE reportCount=reportCount+1, deleted=0", ipdata.IP)
	if err != nil {
		return
	}
	if ce == 0 {
		doAnalytics(ipdata)
	}
	err = queryRow(&IPid, "SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?", ipdata.IP)
	if err != nil {
		return
	}
	reportID = c.ID
	return
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
		return -1
	}
	return uid
}

func isConnectedToDB() error {
	sqlCheckConnection := "SELECT COUNT(*) FROM User"
	var count int
	err := queryRow(&count, sqlCheckConnection)
	if err != nil {
		return err
	}
	return nil
}
