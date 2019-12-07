package main

const (
	//FetchIPs user can fetch IPs
	FetchIPs = int16(1) << iota
	//PushIPs  can push IPs
	PushIPs = int16(1) << iota
	//ViewReports user can view reports
	ViewReports = int16(1) << iota
)

func canUser(permission int16, flag int16) bool {
	return permission&flag == flag
}
