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

	initDB(readConfig("credentials.json"))

	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))

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
