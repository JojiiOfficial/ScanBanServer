package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/mkideal/cli"
	"github.com/thecodeteam/goodbye"
)

type fixIpdataT struct {
	cli.Helper
	StartID             uint `cli:"*s,start" usage:"Specify the start ID to fix from"`
	UpdateAlreadyFilled bool `cli:"f,update-filled" usage:"Update IPs having type >= 0 instead of type=0"`
}

var fixIpdataCMD = &cli.Command{
	Name:    "fixipdata",
	Aliases: []string{},
	Desc:    "fixes missing ipdata in DB",
	Argv:    func() interface{} { return new(fixIpdataT) },
	Fn: func(ct *cli.Context) error {
		argv := ct.Argv().(*fixIpdataT)
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

		initDB(config)
		dberr := isConnectedToDB()
		if dberr != nil {
			LogError("Couldn't connect to database: " + dberr.Error())
			return nil
		}

		add := "type = 0"
		if argv.UpdateAlreadyFilled {
			add = "type >=0"
		}
		success := connectIPDataClient(config)
		if !success {
			return errors.New("Couldn't connect to ipdata.co")
		}

		var ids []string
		err = queryRows(&ids, "SELECT ip FROM BlockedIP WHERE "+add+" AND pk_id >= "+strconv.FormatUint(uint64(argv.StartID), 10))
		if err != nil {
			return errors.New("error: " + err.Error())
		}
		for i, ip := range ids {
			percent := int(i * 100 / len(ids))
			fmt.Println(strconv.Itoa(percent) + "% IP:" + ip)
			runAnalytic(ip)
		}
		fmt.Println("Done")

		return nil
	},
}
