package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/thecodeteam/goodbye"
)

func main() {
	ctx := context.Background()
	defer goodbye.Exit(ctx, -1)
	goodbye.Notify(ctx)
	goodbye.Register(func(ctx context.Context, sig os.Signal) {
		if db != nil {
			_ = db.Close()
			fmt.Println("DB closed")
		}
	})

	_, err := os.Stat("./credentials.json")
	if err != nil {
		fmt.Println("Couldn't find credentials.json")
		return
	}

	fmt.Println("Requesting ip")
	ip, errcode := retrieveExternIP()
	if errcode != 0 {
		return
	}
	fmt.Println("Your external ip is: " + ip)

	initDB(readConfig("credentials.json"))

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
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
