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
		LogInfo("Duration: " + time.Since(start).String())
	})
}
