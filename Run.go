package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mkideal/cli"
	"github.com/theckman/go-ipdata"
	"github.com/thecodeteam/goodbye"
)

type runT struct {
	cli.Helper
}

var ipdataClient *ipdata.Client
var config Config
var filterprocessor Filterprocessor

var runCMD = &cli.Command{
	Name:    "run",
	Aliases: []string{},
	Desc:    "run the server service",
	Argv:    func() interface{} { return new(runT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*runT)
		_ = argv
		ctx := context.Background()
		defer goodbye.Exit(ctx, -1)
		goodbye.Notify(ctx)
		goodbye.Register(func(ctx context.Context, sig os.Signal) {
			if db != nil {
				_ = db.Close()
				LogInfo("DB closed")
			}
		})

		_, err := os.Stat("./config.json")
		if err != nil {
			LogError("Couldn't find config.json")
			return nil
		}

		config = readConfig("config.json")
		showTimeInLog = config.ShowTimeInLog

		_, err = os.Stat("./dyn.ip")
		useDynDNS = false
		if err != nil {
			if !handleIPRequest() {
				return nil
			}
		} else {
			LogInfo("Found dyn.ip file. Trying to use it")
			ip, err := ioutil.ReadFile("./dyn.ip")
			if err != nil {
				LogError("Couldn't use dyn.ip file! Using extern ip")
				if !handleIPRequest() {
					return nil
				}
			} else {
				ipe := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(string(ip), "\n", ""), "\r", ""), " ")
				valid, r := isIPValid(ipe)
				if !valid {
					if r == -1 {
						LogError("IP is a reserved ip!")
					}
					LogError("You got ip: " + ipe + " but its not valid!")
					if !handleIPRequest() {
						return nil
					}
					LogError("Using a static ip!")
				} else {
					LogInfo("Your ip is: " + ipe)
					useDynDNS = true
					externIP = ipe
				}
			}
		}

		initDB(config)
		dberr := isConnectedToDB()
		if dberr != nil {
			LogError("Couldn't connect to database: " + dberr.Error())
			return nil
		}

		useTLS := false
		if len(config.CertFile) > 0 {
			_, err = os.Stat(config.CertFile)
			if err != nil {
				LogError("Certfile not found. HTTP only!")
				useTLS = false
			} else {
				useTLS = true
			}
		}

		if len(config.KeyFile) > 0 {
			_, err = os.Stat(config.KeyFile)
			if err != nil {
				LogError("Keyfile not found. HTTP only!")
				useTLS = false
			}
		}
		connectIPDataClient(config)
		router := NewRouter()
		if useTLS {
			go (func() {
				log.Fatal(http.ListenAndServeTLS(":8081", config.CertFile, config.KeyFile, router))
			})()
		}

		filterprocessor.start()
		log.Fatal(http.ListenAndServe(":8080", router))

		return nil
	},
}

func connectIPDataClient(config Config) {
	if len(config.IPdataAPIKey) > 0 {
		ipd, err := ipdata.NewClient(config.IPdataAPIKey)
		if err != nil {
			LogError("Could not connect to IPdata.co: " + err.Error())
			ipdataClient = nil
		} else {
			LogInfo("Successfully connected to Ipdata.co")
			ipdataClient = &ipd
		}
	}
}
