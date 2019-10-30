package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/mkideal/cli"
	"github.com/thecodeteam/goodbye"
)

type runT struct {
	cli.Helper
}

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
				fmt.Println("DB closed")
			}
		})

		_, err := os.Stat("./config.json")
		if err != nil {
			fmt.Println("Couldn't find config.json")
			return nil
		}

		_, err = os.Stat("./dyn.ip")
		useDynDNS = false
		if err != nil {
			if !handleIPRequest() {
				return nil
			}
		} else {
			fmt.Println("Found dyn.ip file. Trying to use it")
			ip, err := ioutil.ReadFile("./dyn.ip")
			if err != nil {
				fmt.Println("Couldn't use dyn.ip file! Using extern ip")
				if !handleIPRequest() {
					return nil
				}
			} else {
				ipe := strings.Trim(strings.ReplaceAll(strings.ReplaceAll(string(ip), "\n", ""), "\r", ""), " ")
				valid, r := isIPValid(ipe)
				if !valid {
					if r == -1 {
						fmt.Println("IP is a reserved ip!")
					}
					fmt.Println("You got ip:", ipe, "but its not valid!")
					if !handleIPRequest() {
						return nil
					}
					fmt.Println("Using a static ip!")
				} else {
					fmt.Println("Your ip is: " + ipe)
					useDynDNS = true
					externIP = ipe
				}
			}
		}
		config := readConfig("config.json")
		initDB(config)
		useTLS := true
		_, err = os.Stat(config.CertFile)
		if err != nil {
			fmt.Println("Certfile not found. HTTP only!")
			useTLS = false
		}
		_, err = os.Stat(config.KeyFile)
		if err != nil {
			fmt.Println("Keyfile not found. HTTP only!")
			useTLS = false
		}

		router := NewRouter()
		if useTLS {
			go (func() {
				log.Fatal(http.ListenAndServeTLS(":8081", config.CertFile, config.KeyFile, router))
			})()

		}
		log.Fatal(http.ListenAndServe(":8080", router))

		return nil
	},
}
