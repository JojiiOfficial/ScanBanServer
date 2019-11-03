package main

import (
	"log"
	"net/http"
	"time"
)

//Logger logs stuff
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		go log.Printf(
			"%s\t%s\t%s\t%s",
			r.Method,
			r.RequestURI,
			name,
			time.Since(start),
		)
	})
}

const (
	//Info log
	Info = 1
	//LogError error log
	LogError = 2
	//Critical critical error log
	Critical = 3
)

func logTypeToString(logType int) string {
	switch logType {
	case Info:
		{
			return "[info]"
		}
	case LogError:
		{
			return "[!error!]"
		}
	case Critical:
		{
			return "[*!Critical!*]"
		}
	default:
		return "[ ]"
	}
}

//PrintLogError prints error with given
func PrintLogError(logType int, message string) {
	log.Printf(
		"%s\t%s",
		logTypeToString(logType),
		message,
	)
}
