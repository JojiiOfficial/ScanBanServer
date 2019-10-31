package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/mkideal/cli"
)

var help = cli.HelpCommand("display help information")

type argT struct {
	cli.Helper
}

var root = &cli.Command{
	Argv: func() interface{} { return new(argT) },
	Fn: func(ctx *cli.Context) error {
		fmt.Println("Usage: scanban <install/disable/start/stop>")
		return nil
	},
}

func main() {
	if err := cli.Root(root,
		cli.Tree(help),
		cli.Tree(runCMD),
		cli.Tree(installCMD),
		cli.Tree(stopCMD),
		cli.Tree(startCMD),
	).Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var externIP string
var useDynDNS bool
var serviceName = "ScanBanServer"

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
	valid, _ := isIPValid(ipe)
	if !valid {
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
	valid, _ := isIPValid(ipe)
	if !valid {
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
