package main

import "time"

// ------------- Database structs ----------------

//User user in db
type User struct {
	Pkid        uint      `db:"pk_id"`
	MachineName string    `db:"machineName"`
	Token       string    `db:"token"`
	ReportedIPs uint      `db:"reportedIPs"`
	LastReport  time.Time `db:"lastReport"`
	CreatedAt   time.Time `db:"createdAt"`
	IsValid     bool      `db:"isValid"`
}

// -------------- REST structs ----------------------

//ReportStruct report ips data
type ReportStruct struct {
	Token     string   `json:"tk"`
	StartTime uint64   `json:"st"`
	IPs       []IPData `json:"ips"`
}

//IPData ipdata for Reportstruct
type IPData struct {
	IP    string         `json:"ip"`
	Ports []IPPortReport `json:"prt"`
}

//IPPortReport reportdata for one ip
type IPPortReport struct {
	Port  int   `json:"p"`
	Times []int `json:"t"`
}

//ReportIPStruct incomming ip report
type ReportIPStruct struct {
	Token string  `json:"token"`
	Note  *string `json:"note,omitempty"`
	Ips   []IPset `json:"ips"`
}

//Status a REST response status
type Status struct {
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
}

//PingRequest request strct for fetching changed ips
type PingRequest struct {
	Token string `json:"token"`
}

//FetchRequest request strct for fetching changed ips
type FetchRequest struct {
	Token  string      `json:"token"`
	Filter FetchFilter `json:"filter"`
}

//FetchFilter to filter result from fetch request
type FetchFilter struct {
	Since            int64   `json:"since"`
	MinReason        float64 `json:"minReason"`
	MinReports       int     `json:"minReports"`
	ProxyAllowed     int     `json:"allowProxy"`
	MaxIPs           uint    `json:"maxIPs"`
	OnlyValidatedIPs int     `json:"onlyValid"`
}

//FetchResponse struct for fetch response
type FetchResponse struct {
	IPs              []IPList `json:"ips"`
	CurrentTimestamp int64    `json:"cts"`
}

// -------------- Datatypes structs ----------------------

//IPset a report set containing ip and a reason
type IPset struct {
	IP     string `json:"ip"`
	Reason int    `json:"r"`
	Valid  int    `json:"v"`
}

//IPList a list of ips from DB
type IPList struct {
	IP      string `db:"ip" json:"ip"`
	Deleted int    `db:"deleted" json:"del"`
}

//IPInfoRequest request for ipinfo
type IPInfoRequest struct {
	Token string   `json:"t"`
	IPs   []string `json:"ips"`
}

//IPInfoData data for IPInfo
type IPInfoData struct {
	IP      string       `json:"ip"`
	Reports []ReportData `json:"reports"`
}

//ReportData data for a report
type ReportData struct {
	ReporterID   int    `json:"repid" db:"reporterID"`
	ReporterName string `json:"repnm" db:"machineName"`
	Time         int64  `json:"tm" db:"scanDate"`
	Port         int    `json:"prt" db:"port"`
	Count        int    `json:"ct" db:"count"`
}

//UserPermissions permissions for user
type UserPermissions struct {
	UID         uint  `db:"pk_id"`
	Permissions int16 `db:"permissions"`
}

//Filter a filterobject from database
type Filter struct {
	ID   uint `db:"pk_id"`
	Rows []FilterRow
	Skip bool
}

//FilterRow a row in filter
type FilterRow struct {
	Row []*FilterPart
}

//FilterRowRaw raw row data from db
type FilterRowRaw struct {
	ID        uint  `db:"pk_id"`
	FilterID  uint  `db:"filterID"`
	RowNumber uint8 `db:"rowNumber"`
	PartID    uint  `db:"partID"`
}

//FilterPart part of a filter
type FilterPart struct {
	ID       uint   `db:"pk_id"`
	Dest     uint8  `db:"dest"`
	Operator uint8  `db:"operator"`
	Val      string `db:"val"`
}

//InitNewfilterRequest request for init new filter
type InitNewfilterRequest struct {
	AuthToken string `json:"token"`
	FilterID  uint   `json:"filterID"`
}

//UpdateFilterCacheRequest request for init new filter
type UpdateFilterCacheRequest struct {
	AuthToken string `json:"token"`
}

//FilterTokenCount count and filterID
type FilterTokenCount struct {
	FilterID    uint `db:"filter"`
	FilterCount uint `db:"c"`
}
