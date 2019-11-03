package main

import (
	"log"
	"os"
)

//DefaultLogger logs os.StdOut
var DefaultLogger *log.Logger

//ErrorLogger logs os.StdErr
var ErrorLogger *log.Logger

func initLogger(prefix string) {
	timeFlag := log.Ldate | log.Ltime
	if !showTimeInLog {
		timeFlag = 0
	}
	if DefaultLogger == nil {
		DefaultLogger = log.New(os.Stdout, prefix, timeFlag)
	}
	if ErrorLogger == nil {
		ErrorLogger = log.New(os.Stderr, prefix, timeFlag)
	}
}

//Log logs
func Log(level int, errLogger bool, msg string) {
	initLogger(logPrefix)

	logg := DefaultLogger
	if errLogger {
		logg = ErrorLogger
	}

	logg.Printf(
		"%s %s",
		logTypeToString(level),
		msg,
	)
}

//LogCritical logs a very critical error
func LogCritical(msg string) {
	Log(LogCrit, true, msg)
}

//LogError logs error message
func LogError(msg string) {
	Log(LogErr, true, msg)
}

//LogInfo logs info message
func LogInfo(msg string) {
	Log(LogInf, false, msg)
}

//PrintFInfo fprints info
func PrintFInfo(format string, data ...interface{}) {
	initLogger(logPrefix)
	go DefaultLogger.Printf(format, data...)
}

const (
	//LogInf log
	LogInf = 1
	//LogErr error log
	LogErr = 2
	//LogCrit critical error log
	LogCrit = 3
)

func logTypeToString(logType int) string {
	switch logType {
	case LogInf:
		{
			return "[info]"
		}
	case LogErr:
		{
			return "[!error!]"
		}
	case LogCrit:
		{
			return "[*!Critical!*]"
		}
	default:
		return "[ ]"
	}
}
