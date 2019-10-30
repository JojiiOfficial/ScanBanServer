package main

import (
	"GoSystemd/SystemdGoService"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/mkideal/cli"
)

type installT struct {
	cli.Helper
}

var installCMD = &cli.Command{
	Name:    "install",
	Aliases: []string{},
	Desc:    "installst the server",
	Argv:    func() interface{} { return new(installT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*installT)
		_ = argv
		if SystemdGoService.SystemfileExists(serviceName) {
			fmt.Println("Service already exists")
			return nil
		}

		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		service := SystemdGoService.NewDefaultService(serviceName, "The legendary IP database server", ex+" run")
		service.Service.User = "root"
		service.Service.Group = "root"
		service.Service.Restart = SystemdGoService.Always
		cpath, _ := filepath.Abs(ex)
		cpath, _ = path.Split(cpath)
		service.Service.WorkingDirectory = cpath
		service.Service.RestartSec = "3"
		err = service.Create()
		if err == nil {
			service.Enable()
			service.Start()
			fmt.Println("Service installed and started")
		} else {
			fmt.Println("An error occured installitg the service: ", err.Error())
		}

		return nil
	},
}
