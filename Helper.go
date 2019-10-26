package main

import (
	"strconv"
	"strings"
)

func isIPValid(ip string) bool {
	_, err := strconv.Atoi(strings.ReplaceAll(ip, ",", ""))
	return err != nil
}
