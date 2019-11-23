package main

import "time"

// ------------- Database structs ----------------

//User user in db
type User struct {
	Pkid        int       `db:"pk_id"`
	Username    string    `db:"username"`
	Token       string    `db:"token"`
	ReportedIPs int       `db:"reportedIPs"`
	LastReport  time.Time `db:"lastReport"`
	CreatedAt   time.Time `db:"createdAt"`
	IsValid     bool      `db:"isValid"`
}

// -------------- REST structs ----------------------

//ReportStruct report ips data
type ReportStruct struct {
	Token     string   `json:"tk"`
	StartTime int64    `json:"st"`
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

//IPID a pair of an IP in db with its ID
type IPID struct {
	ID int    `db:"pk_id"`
	IP string `db:"ip"`
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
	ReporterName string `json:"repnm" db:"username"`
	Time         int64  `json:"tm" db:"scanDate"`
	Port         int    `json:"prt" db:"port"`
	Count        int    `json:"ct" db:"count"`
}
