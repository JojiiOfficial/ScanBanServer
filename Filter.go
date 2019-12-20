package main

import (
	"strconv"
	"strings"
)

func operatorToSQL(operator uint8) string {
	switch operator {
	case 1:
		return "="
	case 2:
		return "<"
	case 3:
		return ">"
	case 4:
		return "!="
	case 5:
		return "IN"
	case 6:
		return "NOT IN"
	default:
		return ""
	}
}

func destToSQL(dest uint8) string {
	switch dest {
	case 1:
		return "reportCount"
	case 2:
		return "isProxy"
	case 3:
		return "validated"
	case 4:
		return "lastReport"
	case 5:
		return "firstReport"
	case 6:
		return "domain"
	case 7:
		return "hostname"
	case 8:
		return "type"
	case 9:
		return "KnownAbuser"
	case 10:
		return "KnownHacker"
	case 11:
		return "IPports.port"
	case 12:
		return "(SELECT sum(count) FROM `ReportPorts` JOIN Report on Report.pk_id = ReportPorts.reportID WHERE ip=? GROUP BY ip, MINUTE(FROM_UNIXTIME(scanDate)), DATE(FROM_UNIXTIME(scanDate)) ORDER BY SUM(count) DESC LIMIT 1)"
	case 13:
		return "(SELECT sum(count) FROM `ReportPorts` JOIN Report on Report.pk_id = ReportPorts.reportID WHERE ip=? GROUP BY ip, HOUR(FROM_UNIXTIME(scanDate)), DATE(FROM_UNIXTIME(scanDate)) ORDER BY SUM(count) DESC LIMIT 1)"
	default:
		return ""
	}
}

func filterPartToSQL(part FilterPart, ip string) string {
	operator := operatorToSQL(part.Operator)
	column := destToSQL(part.Dest)
	if part.Dest == 12 || part.Dest == 13 {
		column = strings.ReplaceAll(column, "?", ip)
	}
	if len(operator) == 0 {
		return ""
	}
	if len(column) == 0 {
		return ""
	}
	val := part.Val
	if part.Dest == 2 || part.Dest == 3 || part.Dest == 9 || part.Dest == 10 {
		if val == "true" {
			val = "1"
		} else if val == "false" {
			val = "0"
		} else {
			LogError("wrog bool value \"" + val + "\" for part: " + strconv.FormatUint(uint64(part.ID), 10))
			return ""
		}
	}
	if part.Operator == 5 || part.Operator == 6 {
		if part.Dest == 6 || part.Dest == 7 {
			d := strings.Split(part.Val, ",")
			e := ""
			for _, s := range d {
				e += "'" + s + "',"
			}
			e = e[:len(e)-1]
			val = e
		} else if part.Dest != 8 && part.Dest != 11 {
			return ""
		}
		val = "(" + val + ")"
	} else if !isNumeric(val) {
		val = "'" + val + "'"
	}
	return column + " " + operator + " " + val
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func getFilterSQL(filter Filter, ip string) (string, string, error) {
	sqlWhere := ""
	joinAdd3 := ""
	for _, row := range filter.Rows {
		matchRow, joinAdd, err := getFilterRowSQL(row, ip)
		if err != nil || len(strings.Trim(matchRow, " ")) == 0 {
			return "", "", err
		}
		sqlWhere += "(" + matchRow + ") OR"
		if len(joinAdd) > 0 {
			joinAdd3 += joinAdd
		}
	}
	if strings.HasSuffix(sqlWhere, "OR") {
		sqlWhere = sqlWhere[:len(sqlWhere)-3]
	}
	return sqlWhere, joinAdd3, nil
}

func getFilterRowSQL(rowData FilterRow, ip string) (string, string, error) {
	rowSQL := ""
	joinAdd := ""
	for _, row := range rowData.Row {
		part := filterPartToSQL(*row, ip)
		if len(part) > 0 {
			if len(rowData.Row) > 1 {
				part = "(" + part + ")"
			}
			rowSQL += part + " AND"
			if row.Dest == 11 {
				joinAdd += " JOIN IPports on IPports.ip = BlockedIP.pk_id "
			}
		}
	}
	if strings.HasSuffix(rowSQL, "AND") {
		rowSQL = rowSQL[:len(rowSQL)-4]
	}
	return rowSQL, joinAdd, nil
}
