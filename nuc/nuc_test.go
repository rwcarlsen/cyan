package nuc

import (
	"math"
	"testing"
)

func TestFPE(t *testing.T) {
	m := Material{
		922350000: 1,
	}

	fpe := FPE(m)
	nmols := 1000.0 / 235.0
	expected := nmols * Mol * 200 * MeV

	if math.Abs(fpe-expected) > 1e-6 {
		t.Errorf("fpe: expected %v J, got %v J", expected, fpe)
	}
}
