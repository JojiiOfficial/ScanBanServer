package main

import (
	"GoSystemd/SystemdGoService"
	"fmt"
	"os/exec"

	"github.com/mkideal/cli"
)

type startT struct {
	cli.Helper
}

var startCMD = &cli.Command{
	Name:    "run",
	Aliases: []string{},
	Desc:    "starts the server",
	Argv:    func() interface{} { return new(startT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*startT)
		_ = argv
		if !SystemdGoService.SystemfileExists("ScanBanServer") {
			fmt.Println("No server found. Use ./scanban install to install it")
			return
		}
		return nil
	},
}

func runCommand(errorHandler func(error, string), sCmd string) (outb string, err error) {
	out, err := exec.Command("su", "-c", sCmd).Output()
	output := string(out)
	if err != nil {
		if errorHandler != nil {
			errorHandler(err, sCmd)
		}
		return "", err
	}
	return output, nil
}
