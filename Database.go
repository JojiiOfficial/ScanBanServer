package main

import (
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

//DBConfig config for DB
type DBConfig struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Pass     string `json:"pass"`
	Port     int    `json:"port"`
}

var db *sqlx.DB

func initDB(config DBConfig) {
	var err error
	db, err = sqlx.Open("mysql", config.Username+":"+config.Pass+"@tcp("+config.Host+":"+strconv.Itoa(config.Port)+")/"+config.Username)
	if err != nil {
		panic(err)
	}
}

func queryRow(a interface{}, query string, args ...interface{}) error {
	err := db.Get(a, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func queryRows(a interface{}, query string, args ...interface{}) error {
	err := db.Select(a, query, args...)
	if err != nil {
		return err
	}
	return nil
}

func execDB(query string, args ...interface{}) error {
	tx := db.MustBegin()
	tx.MustExec(query, args...)
	err := tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func namedExecDB(query string, arg interface{}) error {
	tx := db.MustBegin()
	_, err := tx.NamedExec(query, arg)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
