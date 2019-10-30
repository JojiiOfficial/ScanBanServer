package main

import (
	"GoSystemd/SystemdGoService"
	"fmt"

	"github.com/mkideal/cli"
)

type stopT struct {
	cli.Helper
}

var stopCMD = &cli.Command{
	Name:    "stop",
	Aliases: []string{},
	Desc:    "stops the server",
	Argv:    func() interface{} { return new(stopT) },
	Fn: func(ct *cli.Context) error {
		if !SystemdGoService.SystemfileExists(SystemdGoService.NameToServiceFile(serviceName)) {
			fmt.Println("No server found. Use './scanban install' to install it")
			return nil
		}

		err := SystemdGoService.SetServiceStatus(serviceName, SystemdGoService.Stop)
		if err != nil {
			fmt.Println("Error stopping service: ", err.Error())
			return nil
		}
		fmt.Println("stopped successfully")

		return nil
	},
}
