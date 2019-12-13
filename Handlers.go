package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"time"
)

func fetchIPInfo(w http.ResponseWriter, r *http.Request) {
	var ipinforequest IPInfoRequest
	if !handleUserInput(w, r, &ipinforequest) {
		return
	}
	if isStructInvalid(ipinforequest) {
		sendError("input missing", w, WrongInputFormatError, 422)
		return
	}
	if len(ipinforequest.Token) != 64 {
		sendError("wrong token length", w, InvalidTokenError, 422)
		return
	}

	ips := []string{}
	validIPfound := false
	ownIP := getOwnIP()
	for _, ip := range ipinforequest.IPs {
		if valid, reas := isIPValid(ip); valid && ip != ownIP {
			ips = append(ips, ip)
			validIPfound = true
		} else {
			add := ""
			if reas != 1 {
				if reas == 0 {
					add = "No valid ipv4"
				} else if reas == -1 {
					add = "IP is reserved"
				}
			} else if ip == ownIP {
				add = "IP is servers IP"
			}
			LogInfo("IP \"" + ip + "\" is not valid! " + add)
		}
	}

	if validIPfound {
		c, data := getIPInfo(ips, ipinforequest.Token)
		if c == -1 {
			sendError("User invalid", w, InvalidTokenError, 422)
			return
		} else if c == 2 {
			sendError("Server error", w, ServerError, 422)
			return
		} else if c == -3 {
			sendError("No permission", w, NoPermissionError, 403)
			return
		}
		handleError(sendSuccess(w, data), w, ServerError, 500)
	} else {
		sendError("no valid ip found in report", w, NoValidIPFound, 422)
		return
	}
}

func reportIPs(w http.ResponseWriter, r *http.Request) {
	var report ReportStruct

	if !handleUserInput(w, r, &report) {
		return
	}
	if isStructInvalid(report) {
		sendError("input missing", w, WrongInputFormatError, 422)
		return
	}

	repIP := getIPFromHTTPrequest(r)

	ips := []IPData{}
	validIPfound := false
	ownIP := getOwnIP()
	for _, ip := range report.IPs {
		sIP := ip.IP
		if valid, reas := isIPValid(sIP); valid && sIP != repIP && sIP != ownIP {
			ips = append(ips, ip)
			validIPfound = true
		} else {
			add := ""
			if sIP == repIP {
				add = "IP is reporters IP"
			} else if reas != 1 {
				if reas == 0 {
					add = "No valid ipv4"
				} else if reas == -1 {
					add = "IP is reserved"
				}
			} else if sIP == ownIP {
				add = "IP is servers IP"
			}
			LogInfo("IP \"" + sIP + "\" is not valid! " + add)
		}
	}

	if validIPfound {
		c := insertIPs(report.Token, ips, report.StartTime)
		if c == -1 {
			sendError("User invalid", w, InvalidTokenError, 422)
			return
		} else if c == 2 {
			sendError("Server error", w, ServerError, 422)
			return
		} else if c == -3 {
			sendError("No permission", w, NoPermissionError, 403)
			return
		}
		handleError(sendSuccess(w, Status{
			StatusCode:    "success",
			StatusMessage: "success",
		}), w, ServerError, 500)
	} else {
		sendError("no valid ip found in report", w, NoValidIPFound, 422)
		return
	}
}

func fetchIPs(w http.ResponseWriter, r *http.Request) {
	var fetchRequest FetchRequest

	if !handleUserInput(w, r, &fetchRequest) {
		return
	}

	if isStructInvalid(fetchRequest) {
		sendError("input missing", w, WrongInputFormatError, 422)
		return
	}

	if len(fetchRequest.Token) != 64 {
		sendError("wrong token length", w, InvalidTokenError, 422)
		return
	}

	ips, err := fetchIPsFromDB(fetchRequest.Token, fetchRequest.Filter)
	if err == -1 {
		sendError("User invalid", w, InvalidTokenError, 422)
		return
	} else if err == -2 {
		sendError("Server error", w, ServerError, 422)
		return
	} else if err == -3 {
		sendError("No permission", w, NoPermissionError, 403)
		return
	} else if err >= 0 {
		if len(ips) == 0 {
			handleError(sendSuccess(w, []string{}), w, ServerError, 500)
		} else {
			fetchresponse := FetchResponse{
				IPs:              ips,
				CurrentTimestamp: time.Now().Unix(),
			}
			handleError(sendSuccess(w, fetchresponse), w, ServerError, 500)
		}
	} else {
		sendError("Server error", w, ServerError, 422)
	}
}

func handleUserInput(w http.ResponseWriter, r *http.Request, p interface{}) bool {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 10000))
	if err != nil {
		LogError("ReadError: " + err.Error())
		return false
	}
	if err := r.Body.Close(); err != nil {
		LogError("ReadError: " + err.Error())
		return false
	}

	errEncode := json.Unmarshal(body, p)
	if handleError(errEncode, w, WrongInputFormatError, 422) {
		return false
	}
	return true
}

func handleError(err error, w http.ResponseWriter, message ErrorMessage, statusCode int) bool {
	if err == nil {
		return false
	}
	sendError(err.Error(), w, message, statusCode)
	return true
}

func sendError(erre string, w http.ResponseWriter, message ErrorMessage, statusCode int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if statusCode >= 500 {
		LogCritical(erre)
	} else {
		LogError(erre)
	}
	w.WriteHeader(statusCode)

	var de []byte
	var err error
	if len(string(message)) == 0 {
		de, err = json.Marshal(&ResponseError)
	} else {
		de, err = json.Marshal(&Status{"error", string(message)})
	}

	if err != nil {
		panic(err)
	}
	_, _ = fmt.Fprintln(w, string(de))
}

func isStructInvalid(x interface{}) bool {
	s := reflect.TypeOf(x)
	for i := s.NumField() - 1; i >= 0; i-- {
		e := reflect.ValueOf(x).Field(i)

		if isEmptyValue(e) {
			return true
		}
	}
	return false
}

func isEmptyValue(e reflect.Value) bool {
	switch e.Type().Kind() {
	case reflect.String:
		if e.String() == "" || strings.Trim(e.String(), " ") == "" {
			return true
		}
	case reflect.Int:
		{
			return false
		}
	case reflect.Int64:
		{
			return false
		}
	case reflect.Array:
		for j := e.Len() - 1; j >= 0; j-- {
			isEmpty := isEmptyValue(e.Index(j))
			if isEmpty {
				return true
			}
		}
	case reflect.Slice:
		return isStructInvalid(e)
	case reflect.Uintptr:
		{
			return false
		}
	case reflect.Ptr:
		{
			return false
		}
	case reflect.UnsafePointer:
		{
			return false
		}
	case reflect.Struct:
		{
			return false
		}
	case reflect.Uint64:
		{
			return false
		}
	default:
		fmt.Println(e.Type().Kind(), e)
		return true
	}
	return false
}

func sendSuccess(w http.ResponseWriter, i interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	de, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, string(de))
	if err != nil {
		return err
	}
	return nil
}
