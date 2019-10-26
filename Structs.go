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
}
