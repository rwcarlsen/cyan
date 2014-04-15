package nuc

const (
	Kg = 1
	g  = 1e-3 * Kg
)

const (
	Joule = 1
	MeV   = 1.602177e-13 * Joule
)

const (
	Atom = 1
	Mol  = 6.022e23 * Atom
)

// Mass represents a quantity in kg
type Mass float64

// Nuc describes a nuclide in ZZZAAAMMMM format.
type Nuc int

// A returns the atomic mass of a nuclide.
func (n Nuc) A() int {
	return (int(n) / 10000) % 1000
}

// Z returns the atomic number of a nuclide.
func (n Nuc) Z() int {
	return int(n) / 10000000
}

// Atoms returns the number of atoms of nuclide n for mass m.
func Atoms(n Nuc, m Mass) float64 {
	v := float64(n.A()) * g / Mol
	return float64(m) / v
}

type Material map[Nuc]Mass

func (m Material) Mass() (tot Mass) {
	for _, qty := range m {
		tot += qty
	}
	return tot
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
	return energy * MeV
}
