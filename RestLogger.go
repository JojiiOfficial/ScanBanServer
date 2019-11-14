package main

import (
	"net/http"
	"time"
)

//Logger logs stuff
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		LogInfo(r.Method + " " + r.RequestURI + " " + name)
		start := time.Now()
		inner.ServeHTTP(w, r)
		dur := time.Since(start)
		if dur > 500*time.Millisecond && dur < 1500*time.Millisecond {
			LogInfo("Duration: " + dur.String())
		} else if dur > 1500*time.Millisecond && dur < 2500*time.Millisecond {
			LogError("Duration: " + dur.String())
		} else if dur > 2500*time.Millisecond {
			LogCritical("Duration: " + dur.String())
		}
	})
}
