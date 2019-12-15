package main

import (
	"fmt"
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
	processor.ipworker = make(chan IPDataResult, 1)
	go (func() {
		for {
			processor.handleIP(<-processor.ipworker)
		}
	})()
}

func (processor *Filterprocessor) handleIP(ipData IPDataResult) {
	success := processor.updateCachedFilter()
	if !success {
		return
	}
	fmt.Println("Working on: "+ipData.IP+" with ID:", ipData.IPID)
}

func (processor *Filterprocessor) addIP(ipData IPDataResult) {
	fmt.Println("added IP: "+ipData.IP+" with ID:", ipData.IPID)
	processor.ipworker <- ipData
}

func (processor *Filterprocessor) updateCachedFilter() bool {
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
	err = queryRows(&filters, "SELECT pk_id FROM Filter WHERE pk_id > ?", processor.lastFilterID)
	if err != nil {
		LogCritical("Couldn't get newest filter:" + err.Error())
		return false
	}
	if len(filters) > 0 {
		processor.lastFilterID = filters[len(filters)-1].ID
	}

	var rowData []FilterRowRaw
	err = queryRows(&rowData, "SELECT pk_id, filterID, rowNumber, partID FROM FilterRow WHERE pk_id > ?", processor.lastFilterRowID)
	if err != nil {
		LogCritical("Couldn't get newest filterRows: " + err.Error())
		return false
	}
	for _, row := range rowData {
		for fi, filter := range filters {
			if filter.ID == row.FilterID {
				for i, part := range processor.filterParts {
					if part.ID == row.PartID {
						for len(filters[fi].Row) <= int(row.RowNumber) {
							filters[fi].Row = append(filters[fi].Row, FilterRow{})
						}
						filters[fi].Row[row.RowNumber].Row = append(filters[fi].Row[row.RowNumber].Row, &(processor.filterParts[i]))
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

	printDebugFilter(processor.filter)

	return true
}

func printDebugFilter(filter []Filter) {
	for _, filter := range filter {
		fmt.Println("FilterID: ", filter.ID)
		for i, row := range filter.Row {
			fmt.Println("  Row", i)
			for _, r := range row.Row {
				fmt.Println("    ", "ID:", r.ID, "data:", r.Val, r.Operator, r.Dest)
			}
		}
	}
}
