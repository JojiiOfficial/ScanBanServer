package main

import (
	"net/http"
	"time"
)

//Logger logs stuff
func Logger(inner http.Handler, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inner.ServeHTTP(w, r)

		LogInfo(r.Method + " " + r.RequestURI + " " + name + " " + time.Since(start).String())
		//PrintFInfo("%s\t%s\t%s\t%s", r.Method, r.RequestURI, name, time.Since(start))
	})
}
