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
	Name:    "start",
	Aliases: []string{},
	Desc:    "starts the server",
	Argv:    func() interface{} { return new(startT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*startT)
		_ = argv
		if !SystemdGoService.SystemfileExists(SystemdGoService.NameToServiceFile(serviceName)) {
			fmt.Println("No server found. Use './scanban install' to install it")
			return nil
		}

		err := SystemdGoService.SetServiceStatus(serviceName, SystemdGoService.Restart)
		if err != nil {
			fmt.Println("Error restarting service: ", err.Error())
			return nil
		}
		fmt.Println("Restarted successfully")

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
