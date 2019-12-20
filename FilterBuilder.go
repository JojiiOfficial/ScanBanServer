package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

//FilterBuilder applies filter to IP
type FilterBuilder struct {
	filterParts      []FilterPart
	filter           []Filter
	lastFilterPartID uint
	lastFilterRowID  uint
	lastFilterID     uint
	handler          FilterHandler
}

func (builder *FilterBuilder) start() {
	builder.cleanUP()
	builder.updateCachedFilter(true)
	builder.handler = FilterHandler{}
	builder.handler.init(3, builder)
}

func (builder *FilterBuilder) addIP(ipData IPDataResult) {
	builder.handler.addIP(ipData)
}

func (builder *FilterBuilder) cleanUP() {
	execDB("DELETE FROM FilterChange")
}

func (builder *FilterBuilder) handleIP(ipData IPDataResult) {
	start := time.Now()
	_ = start
	success := builder.updateCachedFilter(false)
	if !success {
		return
	}
	for i, filter := range builder.filter {
		if filter.Skip {
			continue
		}
		start1 := time.Now()
		sql := "SELECT 1 FROM BlockedIP "
		hasCache := len(filter.SQLCache)
		if hasCache > 0 {
			sql += filter.SQLCache
		} else {
			sqlwhere, joinAdd, err := getFilterSQL(filter, strconv.FormatUint(uint64(ipData.IPID), 10))
			if err != nil {
				LogError("Error apllying filter: " + err.Error())
				continue
			}
			if len(strings.Trim(sqlwhere, " ")) == 0 {
				continue
			}
			scndPart := joinAdd + " WHERE (" + sqlwhere + ") AND BlockedIP.pk_id = "
			sql += scndPart
			builder.filter[i].SQLCache = scndPart
		}
		fmt.Println("Getting filterSQL took", time.Now().Sub(start1).String())

		start1 = time.Now()
		baseSQL := sql + strconv.FormatUint(uint64(ipData.IPID), 10) + " LIMIT 1"
		fmt.Println(baseSQL)
		var hitFilterI int
		hitFilter := true
		err := queryRow(&hitFilterI, baseSQL)
		if err != nil {
			hitFilter = false
		}
		fmt.Println("applying filter took", time.Now().Sub(start1).String())
		start1 = time.Now()
		var alreadyInIPFilter int
		isInFilter := true
		err = queryRow(&alreadyInIPFilter, "SELECT 1 FROM FilterIP WHERE ip=? AND filterID=? LIMIT 1", ipData.IPID, filter.ID)
		if err != nil {
			isInFilter = false
		}
		fmt.Println("IsAlreadyInFilter took: ", time.Now().Sub(start1).String())
		go (func() {
			if hitFilter {
				if !isInFilter {
					execDB("INSERT INTO FilterIP (ip, filterID, added) VALUES(?,?,(SELECT UNIX_TIMESTAMP()))", ipData.IPID, filter.ID)
				}
			} else if hitFilter && isInFilter {
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
			}
		})()
	}
	LogInfo("Applying filter took " + time.Now().Sub(start).String())
}

func (builder *FilterBuilder) updateCachedFilter(initial bool) bool {
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
				for i, f := range builder.filter {
					if f.ID == del {
						builder.filter[i].Skip = true
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
			for i, f := range builder.filter {
				if f.ID == ad {
					builder.filter[i].Skip = false
					break b
				}
			}
		}

		builder.lastFilterRowID = 0
	}

	var parts []FilterPart
	err := queryRows(&parts, "SELECT pk_id, dest, operator, val FROM FilterPart WHERE pk_id > ?", builder.lastFilterPartID)
	if err != nil {
		LogCritical("Couldn't get newest filterparts: " + err.Error())
		return false
	}
	for _, part := range parts {
		builder.filterParts = append(builder.filterParts, part)
	}
	if len(parts) > 0 {
		builder.lastFilterPartID = parts[len(parts)-1].ID
	}

	var filters []Filter
	err = queryRows(&filters, "SELECT DISTINCT Filter.pk_id FROM Filter JOIN Token on Token.filter = Filter.pk_id WHERE Filter.pk_id > ?", builder.lastFilterID)
	if err != nil {
		LogCritical("Couldn't get newest filter:" + err.Error())
		return false
	}
	if len(filters) > 0 {
		builder.lastFilterID = filters[len(filters)-1].ID
	}

	var rowData []FilterRowRaw
	if initial {
		err = queryRows(&rowData, "SELECT pk_id, filterID, rowNumber, partID FROM FilterRow WHERE pk_id > ?", builder.lastFilterRowID)
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
				for i, part := range builder.filterParts {
					if part.ID == row.PartID {
						for len(filters[fi].Rows) <= int(row.RowNumber) {
							filters[fi].Rows = append(filters[fi].Rows, FilterRow{})
						}
						filters[fi].Rows[row.RowNumber].Row = append(filters[fi].Rows[row.RowNumber].Row, &(builder.filterParts[i]))
						break
					}
				}
				break
			}
		}
	}
	if len(rowData) > 0 {
		builder.lastFilterRowID = rowData[len(rowData)-1].ID
	}

	//Append filter to builder->filter
	for _, filter := range filters {
		if len(filter.Rows) > 0 {
			builder.filter = append(builder.filter, filter)
		}
	}

	go (func() {
		execDB("DELETE FROM FilterChange")
	})()

	return true
}
