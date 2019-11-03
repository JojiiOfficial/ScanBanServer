package main

import (
	"log"
	"os"
)

//DefaultLogger logs os.StdOut
var DefaultLogger *log.Logger

//ErrorLogger logs os.StdErr
var ErrorLogger *log.Logger

//Loglevel a loglevel
type Loglevel int

const (
	//LevelInfo log info
	LevelInfo Loglevel = 0
	//LevelError log error
	LevelError Loglevel = 1
	//LevelPanic log panic
	LevelPanic Loglevel = 2
)

func initLogger(prefix string) {
	if DefaultLogger == nil {
		DefaultLogger = log.New(os.Stdout, prefix, log.Ldate|log.Ltime)
	}
	if ErrorLogger == nil {
		ErrorLogger = log.New(os.Stderr, prefix, log.Ldate|log.Ltime)
	}
}

//Log logs
func Log(level Loglevel, msg string) {
	initLogger(logPrefix)
	if level == LevelInfo {
		DefaultLogger.Println(msg)
	} else {
		ErrorLogger.Println(msg)
	}
}

//LogPanic logs a very critical error
func LogPanic(msg string) {
	Log(LevelPanic, msg)
}

//LogError logs error message
func LogError(msg string) {
	Log(LevelPanic, msg)
}

//LogInfo logs info message
func LogInfo(msg string) {
	Log(LevelInfo, msg)
}
