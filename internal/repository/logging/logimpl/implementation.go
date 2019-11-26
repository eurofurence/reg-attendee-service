package logimpl

import (
	"fmt"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/system"
	"log"
)

type LoggerDefaultImpl struct{
	RequestId string
}

const severityDEBUG = "DEBUG"
const severityINFO = "INFO"
const severityWARN = "WARN"
const severityERROR = "ERROR"
const severityFATAL = "FATAL"

const severityINFOPrintAs = "INFO "
const severityWARNPrintAs = "WARN "

var severityMap = map[string]int{
	severityDEBUG: 1,
	severityINFO: 2,
	severityWARN: 3,
	severityERROR: 4,
	severityFATAL: 5,
}

func isEnabled(severity string) bool {
	configured := config.LoggingSeverity()
	return severityMap[severity] >= severityMap[configured]
}

func (l *LoggerDefaultImpl) print(severity5chars string, v ...interface{}) {
	args := []interface{}{"[", l.RequestId, "] ", severity5chars, " "}
	args = append(args, v...)
	log.Print(args...)
}

func (l *LoggerDefaultImpl) IsDebugEnabled() bool {
	return isEnabled(severityDEBUG)
}

func (l *LoggerDefaultImpl) Debug(v ...interface{}) {
	if isEnabled(severityDEBUG) {
		l.print(severityDEBUG, v...)
	}
}

func (l *LoggerDefaultImpl) Debugf(format string, v ...interface{}) {
	if isEnabled(severityDEBUG) {
		l.print(severityDEBUG, fmt.Sprintf(format, v...))
	}
}

func (l *LoggerDefaultImpl) IsInfoEnabled() bool {
	return isEnabled(severityINFO)
}

func (l *LoggerDefaultImpl) Info(v ...interface{}) {
	if isEnabled(severityINFO) {
		l.print(severityINFOPrintAs, v...)
	}
}

func (l *LoggerDefaultImpl) Infof(format string, v ...interface{}) {
	if isEnabled(severityINFO) {
		l.print(severityINFOPrintAs, fmt.Sprintf(format, v...))
	}
}

func (l *LoggerDefaultImpl) IsWarnEnabled() bool {
	return isEnabled(severityWARN)
}

func (l *LoggerDefaultImpl) Warn(v ...interface{}) {
	if isEnabled(severityWARN) {
		l.print(severityWARNPrintAs, v...)
	}
}

func (l *LoggerDefaultImpl) Warnf(format string, v ...interface{}) {
	if isEnabled(severityWARN) {
		l.print(severityWARNPrintAs, fmt.Sprintf(format, v...))
	}
}

func (l *LoggerDefaultImpl) IsErrorEnabled() bool {
	return isEnabled(severityERROR)
}

func (l *LoggerDefaultImpl) Error(v ...interface{}) {
	if isEnabled(severityERROR) {
		l.print(severityERROR, v...)
	}
}

func (l *LoggerDefaultImpl) Errorf(format string, v ...interface{}) {
	if isEnabled(severityERROR) {
		l.print(severityERROR, fmt.Sprintf(format, v...))
	}
}

func (l *LoggerDefaultImpl) Fatal(v ...interface{}) {
	l.print(severityFATAL, v...)
	system.Exit(1)
}

func (l *LoggerDefaultImpl) Fatalf(format string, v ...interface{}) {
	l.print(severityFATAL, fmt.Sprintf(format, v...))
	system.Exit(1)
}
