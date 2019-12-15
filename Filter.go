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
		return "ReportPorts.port"
	case 12:
		return ""
	default:
		return ""
	}
}

func filterPartToSQL(part FilterPart) string {
	operator := operatorToSQL(part.Operator)
	column := destToSQL(part.Dest)
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

	if !isNumeric(val) {
		val = "'" + val + "'"
	}
	return column + " " + operator + " (" + val + ")"
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func getFilterSQL(filter Filter) (string, bool, error) {
	sqlWhere := ""
	hasReportPart := false
	for _, row := range filter.Rows {
		matchRow, hrp, err := getFilterRowSQL(row)
		if err != nil {
			return "", false, err
		}
		sqlWhere += "(" + matchRow + ") OR"
		if hrp {
			hasReportPart = true
		}
	}
	if strings.HasSuffix(sqlWhere, "OR") {
		sqlWhere = sqlWhere[:len(sqlWhere)-3]
	}
	return sqlWhere, hasReportPart, nil
}

func getFilterRowSQL(rowData FilterRow) (string, bool, error) {
	rowSQL := ""
	hasReportPart := false
	for _, row := range rowData.Row {
		part := filterPartToSQL(*row)
		if len(part) > 0 {
			if len(rowData.Row) > 1 {
				part = "(" + part + ")"
			}
			rowSQL += part + " AND"
			if row.Dest == 11 {
				hasReportPart = true
			}
		}
	}
	if strings.HasSuffix(rowSQL, "AND") {
		rowSQL = rowSQL[:len(rowSQL)-4]
	}
	return rowSQL, hasReportPart, nil
}
