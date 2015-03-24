package gone

import "testing"

func TestId(t *testing.T) {
	want := 922350000
	got := Id("U235")
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestIdFromInt(t *testing.T) {
	want := 922350000
	got := IdFromInt(92235)
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestDecayConst(t *testing.T) {
	want := 2.0
	got := DecayConst(Id("Cs137"))
	if got != want {
		t.Errorf("want %v, got %v", want, got)
	}
}
