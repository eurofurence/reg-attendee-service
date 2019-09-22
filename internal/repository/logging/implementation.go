package logging

import (
	"fmt"
	"log"
	"os"
)

type LoggerDefaultImpl struct{
	RequestId string
}

const severityDEBUG = "DEBUG"
const severityINFO = "INFO "
const severityWARN = "WARN "
const severityERROR = "ERROR"
const severityFATAL = "FATAL"

func (l *LoggerDefaultImpl) print(severity5chars string, v ...interface{}) {
	args := []interface{}{"[", l.RequestId, "] ", severity5chars, " "}
	args = append(args, v...)
	log.Print(args...)
}

func (l *LoggerDefaultImpl) Debug(v ...interface{}) {
	l.print(severityDEBUG, v...)
}

func (l *LoggerDefaultImpl) Debugf(format string, v ...interface{}) {
	l.print(severityDEBUG, fmt.Sprintf(format, v...))
}

func (l *LoggerDefaultImpl) Info(v ...interface{}) {
	l.print(severityINFO, v...)
}

func (l *LoggerDefaultImpl) Infof(format string, v ...interface{}) {
	l.print(severityINFO, fmt.Sprintf(format, v...))
}

func (l *LoggerDefaultImpl) Warn(v ...interface{}) {
	l.print(severityWARN, v...)
}

func (l *LoggerDefaultImpl) Warnf(format string, v ...interface{}) {
	l.print(severityWARN, fmt.Sprintf(format, v...))
}

func (l *LoggerDefaultImpl) Error(v ...interface{}) {
	l.print(severityERROR, v...)
}

func (l *LoggerDefaultImpl) Errorf(format string, v ...interface{}) {
	l.print(severityERROR, fmt.Sprintf(format, v...))
}

func (l *LoggerDefaultImpl) Fatal(v ...interface{}) {
	l.print(severityFATAL, v...)
	os.Exit(1)
}

func (l *LoggerDefaultImpl) Fatalf(format string, v ...interface{}) {
	l.print(severityFATAL, fmt.Sprintf(format, v...))
	os.Exit(1)
}
