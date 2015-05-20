package say

import (
	"log"
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
	if Level <= INFO {
		log.Printf("INFO:"+format, v...)
	}
}

func Verbosef(format string, v ...interface{}) {
	if Level <= VERBOSE {
		log.Printf("VERBOSE:"+format, v...)
	}
}

func Debugf(format string, v ...interface{}) {
	if Level <= DEBUG {
		log.Printf(format, v...)
	}
}
