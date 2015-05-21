package say

import (
	"log"
	"os"
	"strings"
)

var Level = DEBUG

const (
	INFO int = iota
	VERBOSE
	DEBUG
)

func confNames() map[string]int {
	return map[string]int{
		"info":    INFO,
		"verbose": VERBOSE,
		"debug":   DEBUG,
	}
}

func SetLevelWithConfName(name string) {
	level, ok := confNames()[strings.ToLower(name)]
	if !ok {
		level = DEBUG
	}

	Level = level
}

func Infof(format string, v ...interface{}) {
	if Level >= INFO {
		log.Printf("INFO:"+format, v...)
	}
}

func Fatalf(format string, v ...interface{}) {
	Infof(format, v...)
	os.Exit(1)
}

func Verbosef(format string, v ...interface{}) {
	if Level >= VERBOSE {
		log.Printf("VERBOSE:"+format, v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if Level >= DEBUG {
		log.Printf(format, v...)
	}
}

func If(level int, action func()) {
	if Level >= level {
		action()
	}
}

func Alwaysf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
