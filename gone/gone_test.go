package gone

import "testing"

func TestId(t *testing.T) {
	want := Nuc(922350000)
	got := Id("U235")
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestIdFromInt(t *testing.T) {
	want := Nuc(922350000)
	got := IdFromInt(92235)
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestDecayConst(t *testing.T) {
	want := 7.302030826339803e-10
	got := DecayConst(Id("Cs137"))
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestFissProdYield(t *testing.T) {
	want := .06221
	got := FissProdYield(Id("U235"), Id("Cs137"), Thermal)
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}
