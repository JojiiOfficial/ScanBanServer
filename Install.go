package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/JojiiOfficial/SystemdGoService"

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
		_, fe := os.Stat("./config.json")
		if fe != nil {
			fmt.Println("Config doesn't exists. Creating a new config.json...\nPlease fill the config.json with the DB credentials")
			(&Config{
				Host:         "localhost",
				Pass:         "A database pass",
				CertFile:     "",
				KeyFile:      "",
				Port:         3066,
				Username:     "DB username",
				IPdataAPIKey: "",
			}).save("./config.json")

			return nil
		}
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

func (config *Config) save(configFile string) error {
	sConf, err := json.Marshal(config)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(configFile, []byte(string(sConf)), 0600)
	if err != nil {
		return err
	}
	return nil
}
