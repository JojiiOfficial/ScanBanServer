package main

import (
	"errors"
	"strconv"
)

//ReportIPcount ip and count
type ReportIPcount struct {
	Count uint `db:"c"`
	ID    uint `db:"iid"`
}

const batchSize = 30

func insertIPs(token string, ipdatas []IPData, starttime uint64) int {
	valid, uid, permissions := IsUserValid(token)

	if !valid {
		return -1
	}

	if !canUser(permissions, PushIPs) {
		return -3
	}

	sqlUpdateUserReportCount := "UPDATE Token SET reportedIPs=reportedIPs+?, lastReport=now() WHERE pk_id=?"
	err := execDB(sqlUpdateUserReportCount, len(ipdatas), uid)
	if err != nil {
		LogCritical("Error updating token reportedIPs")
		return -2
	}

	for _, ipdata := range ipdatas {
		ipID, reportID, err := insertIP(ipdata, uid)
		_ = ipID
		if err != nil {
			LogCritical("Error inserting ip: " + err.Error())
			continue
		}
		if reportID == 0 {
			reportID, err = insertReport(ipdata, uid)
			if err != nil {
				LogCritical("Error inserting report: " + err.Error())
				continue
			}
		}
		err = execDB("UPDATE Report SET lastReport=(SELECT UNIX_TIMESTAMP()) WHERE pk_id=?", reportID)
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

func insertBatch(batch map[int][]int, reportID uint, port int, startTime uint64) {
	values := ""
	for _, b := range batch {
		scanCount := len(b)
		if scanCount == 0 {
			continue
		}
		var rpID int
		err := queryRow(&rpID, "SELECT IFNULL(MAX(pk_id),-1) FROM ReportPorts WHERE scanDate >= ? AND reportID=? AND port=?", startTime-uint64(batchSize), reportID, port)
		if err != nil {
			LogCritical("Couldn't get reportPorts in current batch: " + err.Error())
			continue
		}

		if rpID > 0 {
			err = execDB("UPDATE ReportPorts SET count=count+? WHERE pk_id=?", scanCount, rpID)
		} else {
			values += "(" + strconv.FormatUint(uint64(reportID), 10) + "," + strconv.Itoa(port) + "," + strconv.Itoa(scanCount) + "," + strconv.FormatUint(startTime, 10) + "),"
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
	valid, _, permissions := IsUserValid(token)
	if !valid {
		return -1, nil
	}

	if !canUser(permissions, ViewReports) {
		return -3, nil
	}

	ipdata := []IPInfoData{}
	for _, ip := range ips {
		var info []ReportData
		err := queryRows(&info, "SELECT Report.reporterID, Token.machineName, ReportPorts.scanDate, ReportPorts.port, ReportPorts.count FROM `Report`"+
			"JOIN BlockedIP on (BlockedIP.pk_id = Report.ip)"+
			"JOIN Token on (Token.pk_id = Report.reporterID)"+
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

	go (func() {
		err := execDB("UPDATE Token SET requests=requests+1 WHERE token=?", token)
		if err != nil {
			LogError("Error updating requests count")
		}
	})()

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

func insertReport(ipdata IPData, uid uint) (uint, error) {
	err := execDB("INSERT INTO Report (ip, reporterID, firstReport) VALUES((SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?),?,(SELECT UNIX_TIMESTAMP()))", ipdata.IP, uid)
	if err != nil {
		return 0, err
	}

	var id uint
	err = queryRow(&id, "SELECT Report.pk_id FROM Report WHERE ip=(SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?) AND reporterID=?", ipdata.IP, uid)
	if err != nil {
		return 0, err
	} else if id == 0 {
		return 0, errors.New("report not found")
	}
	return id, nil
}

func insertIP(ipdata IPData, uid uint) (IPid uint, reportID uint, err error) {
	IPid, reportID = 0, 0
	err = nil

	var c ReportIPcount
	err = queryRow(&c, "SELECT COUNT(*) as c, ifnull(pk_id, 0)as iid FROM Report WHERE reporterID=? AND ip=ifnull((SELECT BlockedIP.pk_id FROM BlockedIP WHERE BlockedIP.ip=?),\"\")", uid, ipdata.IP)
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
	err = execDB("INSERT INTO BlockedIP (ip, validated,firstReport, lastReport) VALUES (?,0,(SELECT UNIX_TIMESTAMP()),(SELECT UNIX_TIMESTAMP())) ON DUPLICATE KEY UPDATE reportCount=reportCount+1, deleted=0, lastReport=(SELECT UNIX_TIMESTAMP())", ipdata.IP)
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
	valid, _, permissions := IsUserValid(token)
	if !valid {
		return nil, -1
	}

	if !canUser(permissions, FetchIPs) {
		return nil, -3
	}

	query :=
		"SELECT ip,deleted " +
			"FROM BlockedIP " +
			"WHERE " +
			"(lastReport >= ? OR firstReport >= ?) "

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

	go (func() {
		err := execDB("UPDATE Token SET requests=requests+1 WHERE token=?", token)
		if err != nil {
			LogError("Error updating requests count")
		}
	})()

	return iplist, 0
}

//IsUserValid returns userid if valid or -1 if invalid
func IsUserValid(token string) (bool, uint, int16) {
	sqlCheckUserValid := "SELECT Token.pk_id, Token.permissions FROM Token WHERE token=? AND Token.isValid=1"
	var uid UserPermissions
	err := queryRow(&uid, sqlCheckUserValid, token)
	if err != nil {
		return false, 0, 0
	}
	return true, uid.UID, uid.Permissions
}

func isConnectedToDB() error {
	sqlCheckConnection := "SELECT COUNT(*) FROM Token"
	var count int
	err := queryRow(&count, sqlCheckConnection)
	if err != nil {
		return err
	}
	return nil
}
