package nuc

/*
#include <stdlib.h>
#include "cnucname.h"
*/
import "C"
import (
	"bytes"
	"fmt"
	"unsafe"
)

var NucDataPath = ""

type Nuc int

func Id(nuc string) (Nuc, error) {
	cs := C.CString(nuc)
	defer C.free(unsafe.Pointer(cs))
	n := Nuc(C.id_str(cs))
	if n < 0 {
		return 0, fmt.Errorf("'%v' is not a valid nuclide", n)
	}
	return n, nil
}

func IdFromInt(nuc int) (Nuc, error) {
	n := Nuc(C.id_int(C.int(nuc)))
	if n < 0 {
		return 0, fmt.Errorf("'%v' is not a valid nuclide", n)
	}
	return n, nil
}

func (n Nuc) Z() int { return int(n) / 10000000 }
func (n Nuc) A() int { return (int(n) / 10000) % 1000 }

func (n Nuc) Name() string {
	cname := C.name(C.int(n))
	name := C.GoString(cname)
	C.free(unsafe.Pointer(cname))
	return name
}

const (
	Kg = 1
	g  = 1e-3 * Kg
)

const (
	Joule = 1
	MeV   = 1.602177e-13 * Joule
	MWh   = 3.6e9
)

const (
	Atom = 1
	Mol  = 6.022e23 * Atom
)

// Mass represents a quantity in kg
type Mass float64

// Atoms returns the number of atoms of nuclide n for mass m.
func Atoms(n Nuc, m Mass) float64 {
	v := float64(n.A()) * g / Mol
	return float64(m) / v
}

type Material map[Nuc]Mass

func (m Material) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "# Material mass=%v. Composition:\n", m.Mass())
	for nuc, qty := range m {
		fmt.Fprintf(&buf, "    %v    %v\n", nuc, qty)
	}
	return buf.String()
}

func (m Material) Mass() (tot Mass) {
	for _, qty := range m {
		tot += qty
	}
	return tot
}

func (m Material) SetMass(v Mass) {
	curr := m.Mass()
	for nuc, qty := range m {
		m[nuc] = qty / curr * v
	}
}

func (m Material) EltMass(anum int) (tot Mass) {
	for nuc, qty := range m {
		if nuc.Z() == anum {
			tot += qty
		}
	}
	return tot
}

// FPE returns the amount of fission potential energy in Joules for the
// material described by m.
func FPE(m Material) (energy float64) {
	for nuc, e := range FissFertE {
		energy += e * Atoms(nuc, m[nuc])
	}
	return energy
}
