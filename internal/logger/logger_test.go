package logger

import (
	"testing"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger("DEBUG")
	if log == nil {
		t.Error("expected non-nil logger")
	}
}
