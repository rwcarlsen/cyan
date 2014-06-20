package nuc

import (
	"fmt"
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

	m1 := Material{
		922350000: .04,
		922380000: .96,
	}

	m2 := Material{
		922350000: 156.729,
		922360000: 102.103,
		922380000: 18280.324,
		932370000: 13.656,
		942380000: 5.043,
		942390000: 106.343,
		942400000: 41.357,
		942410000: 36.477,
		942420000: 15.387,
		952410000: 1.234,
		952430000: 3.607,
		962440000: 0.431,
		962450000: 1.263,
	}

	m3 := Material{
		922350000: 0.0027381,
		922380000: 0.9099619,
		942380000: 0.001746,
		942390000: 0.045396,
		942400000: 0.020952,
		942410000: 0.013095,
		942420000: 0.005238,
	}

	m1.SetMass(1)
	m2.SetMass(1)
	m3.SetMass(1)

	fpe1 := FPE(m1)
	fpe2 := FPE(m2)
	fpe3 := FPE(m3)

	fmt.Printf("fpe fresh u fuel: %v\n", fpe1)
	fmt.Printf("fpe spent u fuel: %v\n", fpe2)
	fmt.Printf("fpe fresh mox fuel: %v\n", fpe3)
}
