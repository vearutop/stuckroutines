package stuckroutines_test

import (
	"bytes"
	"testing"

	"github.com/vearutop/stuckroutines/stuckroutines"
)

func TestNewProcessor(t *testing.T) {
	out := bytes.NewBuffer(nil)

	p := stuckroutines.NewProcessor()
	p.Writer = out
	p.Internal()
	p.Report(stuckroutines.Flags{})

	if !bytes.Contains(out.Bytes(), []byte("persistent goroutine(s) found")) {
		t.Errorf("unexpected output: %s", out.String())
	}
}
