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

//ReportIPStruct incomming ip report
type ReportIPStruct struct {
	Token string  `json:"token"`
	Ips   []IPset `json:"ips"`
}

//Status a REST response status
type Status struct {
	StatusCode    string `json:"statusCode"`
	StatusMessage string `json:"statusMessage"`
}

// -------------- Datatypes structs ----------------------

//IPset a report set containing ip and a reason
type IPset struct {
	IP     string `json:"ip"`
	Reason int    `json:"r"`
}

//IPID a pair of an IP in db with its ID
type IPID struct {
	ID int    `db:"pk_id"`
	IP string `db:"ip"`
}
