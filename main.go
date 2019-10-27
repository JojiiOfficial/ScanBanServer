package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/thecodeteam/goodbye"
)

var externIP string
var useDynDNS bool

func main() {
	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		if db != nil {
			_ = db.Close()
			fmt.Println("DB closed")
		}
	})

	_, err := os.Stat("./credentials.json")
	if err != nil {
		fmt.Println("Couldn't find credentials.json")
		return
	}

	_, err = os.Stat("./dyn.ip")
	useDynDNS = false
	if err != nil {
		if !handleIPRequest() {
			return
		}
	} else {
		fmt.Println("Found dyn.ip file. Trying to use it")
		ip, err := ioutil.ReadFile("./dyn.ip")
		if err != nil {
			fmt.Println("Couldn't use dyn.ip file! Using extern ip")
			if !handleIPRequest() {
				return
			}
		} else {
			ipe := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(string(ip), "\n", ""), "\r", ""), " ")
			if !isIPValid(ipe) {
				fmt.Println("You got ip:", ipe, "but its not valid!")
				if !handleIPRequest() {
					return
				}
				fmt.Println("Using a static ip!")
			} else {
				fmt.Println("Your ip is: " + ipe)
				useDynDNS = true
				externIP = ipe
			}
		}
	}

	fmt.Println("", "", "", getOwnIP())

	initDB(readConfig("credentials.json"))

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
}

func getOwnIP() string {
	if useDynDNS {
		return getDYNIP()
	}
	return externIP
}

func getDYNIP() string {
	ip, err := ioutil.ReadFile("./dyn.ip")
	if err != nil {
		fmt.Println("Couldn't use dyn.ip file! Using ip from the start")
		return externIP
	}
	ipe := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(string(ip), "\n", ""), "\r", ""), " ")
	if !isIPValid(ipe) {
		fmt.Println("You got ip:", ipe, "but its not valid! Using ip from the start")
		return externIP
	}
	externIP = ipe
	return ipe
}

func handleIPRequest() bool {
	fmt.Println("Requesting ip")
	ipe, errcode := retrieveExternIP()
	if errcode != 0 {
		return false
	}
	if !isIPValid(ipe) {
		fmt.Println("You got ip:", ipe, "but its not valid!")
		return false
	}
	fmt.Println("Your external ip is: " + ipe)
	externIP = ipe
	return true
}

func retrieveExternIP() (string, int) {
	rest, err := http.Get("https://ifconfig.me/ip")
	if err != nil {
		fmt.Println("Couldn't retrieve extern ip: " + err.Error())
		return "", -1
	}
	resp, err := ioutil.ReadAll(rest.Body)
	if err != nil {
		fmt.Println("Couldn't retrieve extern ip: " + err.Error())
		return "", -1
	}
	return string(resp), 0
}

func readConfig(file string) DBConfig {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	res := DBConfig{}
	err = json.Unmarshal(dat, &res)
	if err != nil {
		panic(err)
	}
	return res
}
