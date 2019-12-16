package main

import (
	"fmt"
	"strconv"
	"time"
)

//Filterprocessor applies filter to IP
type Filterprocessor struct {
	ipworker         chan IPDataResult
	filterParts      []FilterPart
	filter           []Filter
	lastFilterPartID uint
	lastFilterRowID  uint
	lastFilterID     uint
}

func (processor *Filterprocessor) start() {
	processor.cleanUP()
	processor.updateCachedFilter(true)
	processor.ipworker = make(chan IPDataResult, 1)
	go (func() {
		for {
			processor.handleIP(<-processor.ipworker)
		}
	})()
}

func (processor *Filterprocessor) cleanUP() {
	execDB("DELETE FROM FilterChange")
}

func (processor *Filterprocessor) handleIP(ipData IPDataResult) {
	start := time.Now()
	success := processor.updateCachedFilter(false)
	if !success {
		return
	}
	for i, filter := range processor.filter {
		if filter.Skip {
			continue
		}

		start1 := time.Now()
		sql := "SELECT COUNT(BlockedIP.pk_id) FROM BlockedIP "
		hasCache := len(filter.SqlCache)
		if hasCache > 0 {
			sql += filter.SqlCache
		} else {
			sqlwhere, addReportJoin, err := getFilterSQL(filter)
			if err != nil {
				LogError("Error apllying filter: " + err.Error())
				continue
			}
			addJoin := ""
			if addReportJoin {
				addJoin = " JOIN Report on BlockedIP.pk_id = Report.ip JOIN ReportPorts on Report.pk_id = ReportPorts.reportID "
			}
			scndPart := addJoin + "WHERE (" + sqlwhere + ") AND BlockedIP.pk_id = "
			sql += scndPart
			processor.filter[i].SqlCache = scndPart
		}
		fmt.Println("Getting filterSQL took", time.Now().Sub(start1).String())

		start1 = time.Now()
		baseSQL := sql + strconv.FormatUint(uint64(ipData.IPID), 10)
		fmt.Println(baseSQL)
		var hitFilter int
		err := queryRow(&hitFilter, baseSQL)
		if err != nil {
			LogCritical("Error applying filter(" + strconv.FormatUint(uint64(filter.ID), 10) + "): " + err.Error())
			fmt.Println(baseSQL)
			continue
		}
		fmt.Println("applying filter took", time.Now().Sub(start1).String())
		start1 = time.Now()
		var alreadyInIPFilter int
		err = queryRow(&alreadyInIPFilter, "SELECT COUNT(pk_id) FROM FilterIP WHERE ip=? AND filterID=?", ipData.IPID, filter.ID)
		if err != nil {
			LogCritical("Error checking filter" + strconv.FormatUint(uint64(filter.ID), 10) + ": " + err.Error())
			fmt.Println(baseSQL)
			continue
		}
		fmt.Println("IsAlreadyInFilter took: ", time.Now().Sub(start1).String())

		if hitFilter > 0 {
			if alreadyInIPFilter == 0 {
				start1 := time.Now()
				execDB("INSERT INTO FilterIP (ip, filterID, added) VALUES(?,?,(SELECT UNIX_TIMESTAMP()))", ipData.IPID, filter.ID)
				fmt.Println("Insert into filterIpList took: ", time.Now().Sub(start1).String())
			}
		} else if hitFilter == 0 && alreadyInIPFilter > 0 {
			start1 := time.Now()
			err := execDB("INSERT INTO FilterDelete (ip, tokenID) (SELECT ?,Token.pk_id FROM Token WHERE Token.filter=?)", ipData.IPID, filter.ID)
			if err != nil {
				LogCritical("Error inserting deleted in filterdelete: " + err.Error())
				return
			}
			err = execDB("DELETE FROM FilterIP WHERE ip=? AND filterID=?", ipData.IPID, filter.ID)
			if err != nil {
				LogCritical("Error deleting deleted from FilterIP: " + err.Error())
				return
			}
			fmt.Println("Deleting filterIP took: ", time.Now().Sub(start1).String())
		}
	}
	LogInfo("Applying filter took " + time.Now().Sub(start).String())
}

func (processor *Filterprocessor) addIP(ipData IPDataResult) {
	processor.ipworker <- ipData
}

func (processor *Filterprocessor) updateCachedFilter(initial bool) bool {
	if !initial {
		var add []uint
		var delete []uint

		err := queryRows(&add, "SELECT filterID FROM FilterChange WHERE del=0")
		if err != nil {
			LogCritical("Error getting new filter: " + err.Error())
			return false
		}

		err = queryRows(&delete, "SELECT filterID FROM FilterChange WHERE del=1")
		if err != nil {
			LogCritical("Error getting filterdeletions" + err.Error())
			return false
		}

		if len(delete) > 0 {
			for _, del := range delete {
			a:
				for i, f := range processor.filter {
					if f.ID == del {
						processor.filter[i].Skip = true
						break a
					}
				}
			}
			if len(add) == 0 {
				go (func() {
					execDB("DELETE FROM FilterChange")
				})()
			}
		}

		if len(add) == 0 {
			return true
		}

		for _, ad := range add {
		b:
			for i, f := range processor.filter {
				if f.ID == ad {
					processor.filter[i].Skip = false
					break b
				}
			}
		}

		processor.lastFilterRowID = 0
	}

	var parts []FilterPart
	err := queryRows(&parts, "SELECT pk_id, dest, operator, val FROM FilterPart WHERE pk_id > ?", processor.lastFilterPartID)
	if err != nil {
		LogCritical("Couldn't get newest filterparts: " + err.Error())
		return false
	}
	for _, part := range parts {
		processor.filterParts = append(processor.filterParts, part)
	}
	if len(parts) > 0 {
		processor.lastFilterPartID = parts[len(parts)-1].ID
	}

	var filters []Filter
	err = queryRows(&filters, "SELECT DISTINCT Filter.pk_id FROM Filter JOIN Token on Token.filter = Filter.pk_id WHERE Filter.pk_id > ?", processor.lastFilterID)
	if err != nil {
		LogCritical("Couldn't get newest filter:" + err.Error())
		return false
	}
	if len(filters) > 0 {
		processor.lastFilterID = filters[len(filters)-1].ID
	}

	var rowData []FilterRowRaw
	if initial {
		err = queryRows(&rowData, "SELECT pk_id, filterID, rowNumber, partID FROM FilterRow WHERE pk_id > ?", processor.lastFilterRowID)
	} else {
		err = queryRows(&rowData,
			"SELECT pk_id, FilterRow.filterID, rowNumber, partID FROM FilterRow "+
				"JOIN FilterChange on FilterChange.filterID = FilterRow.filterID WHERE FilterChange.del = 0")
	}
	if err != nil {
		LogCritical("Couldn't get newest filterRows: " + err.Error())
		return false
	}
	for _, row := range rowData {
		for fi, filter := range filters {
			if filter.ID == row.FilterID {
				for i, part := range processor.filterParts {
					if part.ID == row.PartID {
						for len(filters[fi].Rows) <= int(row.RowNumber) {
							filters[fi].Rows = append(filters[fi].Rows, FilterRow{})
						}
						filters[fi].Rows[row.RowNumber].Row = append(filters[fi].Rows[row.RowNumber].Row, &(processor.filterParts[i]))
						break
					}
				}
				break
			}
		}
	}
	if len(rowData) > 0 {
		processor.lastFilterRowID = rowData[len(rowData)-1].ID
	}

	//Append filter to processor->filter
	for _, filter := range filters {
		processor.filter = append(processor.filter, filter)
	}

	go (func() {
		execDB("DELETE FROM FilterChange")
	})()

	return true
}

func printDebugFilter(filter []Filter) {
	for _, filter := range filter {
		fmt.Println("FilterID: ", filter.ID)
		for i, row := range filter.Rows {
			fmt.Println("  Row", i)
			for _, r := range row.Row {
				fmt.Println("    ", "ID:", r.ID, "data:", r.Val, r.Operator, r.Dest)
			}
		}
	}
}
