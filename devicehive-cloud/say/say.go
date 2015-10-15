// this package extends standard "log" functionality.
// there are a few additional logging levels
// (from most to less important): WARN INFO DEBUG TRACE
//
// all of these additional levels use log.Printf as an implementation.
package say


import (
	"log"
	"strings"
)


// additional logging level type
type Level int

const (
	WARN Level = iota
	INFO
	DEBUG
	TRACE
)


// global logging level
var logLevel = INFO


// convert custom logging level to string
// return "UNKNOWN" for unknown levels
func (level Level) String() string {
	switch level {
		case WARN: return "WARN"
		case INFO: return "INFO"
		case DEBUG: return "DEBUG"
		case TRACE: return "TRACE"
	}

	return "UNKNOWN" // by default
}

// parseLevel() parses the logging level from a string
// returns INFO by default
func parseLevel(name string) Level {
	switch strings.ToUpper(name) {
		case "WARN": return WARN
		case "INFO": return INFO
		case "DEBUG": return DEBUG
		case "TRACE": return TRACE
	}

	Warnf("%q is unknown logging level, fallback to INFO", name);
	return INFO
}


// SetLevel() sets global logging level
func SetLevel(level Level) {
	if logLevel != level {
		old := logLevel
		logLevel = level
		Debugf("logging level changed from %q to %q", old, level)
	}
}

// SetLevelByName() sets global logging level using string name
func SetLevelByName(name string) {
	SetLevel(parseLevel(name))
}


// Panic() is the same as log.Panic()
//func Panic(v ...interface{}) {
//	log.Panic(v...)
//}

// Panicf() is the same as log.Panicf()
func Panicf(format string, v ...interface{}) {
	log.Panicf(format, v...)
}


// Fatal() is the same as log.Fatal()
//func Fatal(v ...interface{}) {
//	log.Fatal(v...)
//}

// Fatalf() is the same as log.Fatalf()
func Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v...)
}


// Always() is the same as log.Print()
//func Always(v ...interface{}) {
//	log.Print(v...)
//}

// Alwaysf() is the same as log.Printf()
func Alwaysf(format string, v ...interface{}) {
	log.Printf(format, v...)
}


// Warn() is the same as log.Print() with the "WARN: " prefix
// the message is printed if logging level is greater or equal to WARN
//func Warn(v ...interface{}) {
//	if logLevel >= WARN {
//		log.Print("WARN: ", v...)
//	}
//}

// Warnf() is the same as log.Printf() with the "WARN: " prefix
// the message is printed if logging level is greater or equal to WARN
func Warnf(format string, v ...interface{}) {
	if logLevel >= WARN {
		log.Printf("WARN: " + format, v...)
	}
}


// Info() is the same as log.Print() with the "INFO: " prefix
// the message is printed if logging level is greater or equal to INFO
//func Info(v ...interface{}) {
//	if logLevel >= INFO {
//		log.Print("INFO: ", v...)
//	}
//}

// Infof() is the same as log.Printf() with the "INFO: " prefix
// the message is printed if logging level is greater or equal to INFO
func Infof(format string, v ...interface{}) {
	if logLevel >= INFO {
		log.Printf("INFO: " + format, v...)
	}
}


// Debug() is the same as log.Print() with the "DEBUG: " prefix
// the message is printed if logging level is greater or equal to DEBUG
//func Debug(v ...interface{}) {
//	if logLevel >= DEBUG {
//		log.Print("DEBUG: ", v...)
//	}
//}

// Debugf() is the same as log.Printf() with the "DEBUG: " prefix
// the message is printed if logging level is greater or equal to DEBUG
func Debugf(format string, v ...interface{}) {
	if logLevel >= DEBUG {
		log.Printf("DEBUG: " + format, v...)
	}
}


// Trace() is the same as log.Print() with the "TRACE: " prefix
// the message is printed if logging level is greater or equal to TRACE
//func Trace(v ...interface{}) {
//	if logLevel >= TRACE {
//		log.Print("TRACE: ", v...)
//	}
//}

// Tracef() is the same as log.Printf() with the "TRACE: " prefix
// the message is printed if logging level is greater or equal to TRACE
func Tracef(format string, v ...interface{}) {
	if logLevel >= TRACE {
		log.Printf("TRACE: " + format, v...)
	}
}
