package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mkideal/cli"
	"github.com/thecodeteam/goodbye"
)

type dbTestT struct {
	cli.Helper
}

var databasetestCMD = &cli.Command{
	Name:    "dbtest",
	Aliases: []string{},
	Desc:    "tests db",
	Argv:    func() interface{} { return new(dbTestT) },
	Fn: func(ct *cli.Context) error {
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

		var ips []string
		err = queryRows(&ips, "SELECT ip FROM BlockedIP ORDER BY pk_id")
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		for _, ip := range ips {
			go fmt.Println(IPDataRequest(ip, true))
		}

		return nil
	},
}
