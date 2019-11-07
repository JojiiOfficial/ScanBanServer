package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

//Route for REST
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

//Routes all REST routes
type Routes []Route

//NewRouter create new router
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(false)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	return router
}

var routes = Routes{
	Route{
		"report",
		"POST",
		"/report",
		reportIPs2,
	},
	Route{
		"fetchUpdate",
		"POST",
		"/fetch",
		fetchIPs,
	},
}
