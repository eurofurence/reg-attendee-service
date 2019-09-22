package logimpl

import (
	"testing"
)

var cut = &LoggerDefaultImpl{RequestId: "00000000"}

func TestDebug(t *testing.T) {
	cut.Debug("a", "b", "c")
}

func TestInfo(t *testing.T) {
	cut.Info("a", "b", "c")
}

func TestWarn(t *testing.T) {
	cut.Warn("a", "b", "c")
}

func TestError(t *testing.T) {
	cut.Error("a", "b", "c")
}

func TestDebugf(t *testing.T) {
	cut.Debugf("%s + %s + %s", "a", "b", "c")
}

func TestInfof(t *testing.T) {
	cut.Infof("%s + %s + %s", "a", "b", "c")
}

func TestWarnf(t *testing.T) {
	cut.Warnf("%s + %s + %s", "a", "b", "c")
}

func TestErrorf(t *testing.T) {
	cut.Errorf("%s + %s + %s", "a", "b", "c")
}
