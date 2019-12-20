package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mkideal/cli"
	"github.com/theckman/go-ipdata"
	"github.com/thecodeteam/goodbye"
)

type runT struct {
	cli.Helper
}

var ipdataClient *ipdata.Client
var config Config
var filterbuilder FilterBuilder

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

		filterbuilder.start()
		startGraphUpdater()
		log.Fatal(http.ListenAndServe(":8080", router))

		return nil
	},
}

func connectIPDataClient(config Config) bool {
	if len(config.IPdataAPIKey) > 0 {
		ipd, err := ipdata.NewClient(config.IPdataAPIKey)
		if err != nil {
			LogError("Could not connect to IPdata.co: " + err.Error())
			ipdataClient = nil
		} else {
			LogInfo("Successfully connected to Ipdata.co")
			ipdataClient = &ipd
			return true
		}
	}
	return false
}

func updateGraphCache() {
	execDB("DELETE FROM GraphCache")
	execDB("INSERT INTO GraphCache (graphID, time, value1, value2) (SELECT 1 as graphID, scanDate as hour,COUNT(count) as value1,sum(count) as value2 FROM ReportPorts WHERE DATE_SUB(NOW(), INTERVAL 2 DAY) < DATE(FROM_UNIXTIME(scanDate)) AND scanDate < UNIX_TIMESTAMP() GROUP BY HOUR(from_unixtime(scanDate)), DAY(FROM_UNIXTIME(scanDate)), WEEK(FROM_UNIXTIME(scanDate)), MONTH(FROM_UNIXTIME(scanDate)), YEAR(FROM_UNIXTIME(scanDate))ORDER by scanDate DESC LIMIT 25)")
	execDB("INSERT INTO GraphCache (GraphCache.graphID, GraphCache.value1, GraphCache.value2, GraphCache.time) (SELECT 2, ReportPorts.port, SUM(ReportPorts.count) as count, ReportPorts.port FROM Report JOIN ReportPorts on ReportPorts.reportID = Report.pk_id GROUP BY ReportPorts.port)")
}

func startGraphUpdater() {
	go (func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			updateGraphCache()
			<-ticker.C
		}
	})()
}
