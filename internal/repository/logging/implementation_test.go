package logging

import "testing"

func TestDebug(t *testing.T) {
	NoCtx().Debug("a", "b", "c")
}

func TestInfo(t *testing.T) {
	NoCtx().Info("a", "b", "c")
}

func TestWarn(t *testing.T) {
	NoCtx().Warn("a", "b", "c")
}

func TestError(t *testing.T) {
	NoCtx().Error("a", "b", "c")
}

func TestDebugf(t *testing.T) {
	NoCtx().Debugf("%s + %s + %s", "a", "b", "c")
}

func TestInfof(t *testing.T) {
	NoCtx().Infof("%s + %s + %s", "a", "b", "c")
}

func TestWarnf(t *testing.T) {
	NoCtx().Warnf("%s + %s + %s", "a", "b", "c")
}

func TestErrorf(t *testing.T) {
	NoCtx().Errorf("%s + %s + %s", "a", "b", "c")
}
