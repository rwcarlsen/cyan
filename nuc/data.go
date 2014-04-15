package nuc

// Fissile isltopes
const (
	U235  = 922350000
	U233  = 922330000
	Pu239 = 942390000
	Pu241 = 942410000
	Cu243 = 962430000
	Cu245 = 962450000
	Cu247 = 962470000
	Cf251 = 982510000
)

var FissNuc = []Nuc{
	U235,
	U233,
	Pu239,
	Pu241,
	Cu243,
	Cu245,
	Cu247,
	Cf251,
}

// Fertile isltopes
const (
	Th232 = 902320000
	U234  = 922340000
	U238  = 922380000
	Pu238 = 942380000
	Pu240 = 942400000
)

var FertNuc = []Nuc{
	Th232,
	U234,
	U238,
	Pu238,
	Pu240,
}

// FissFertE contains eventual energy release per fission in MeV for fissile
// and fertile isotopes.
var FissFertE = map[Nuc]float64{
	U235:  200 * MeV,
	U233:  200 * MeV,
	Pu239: 200 * MeV,
	Pu241: 200 * MeV,
	Cu243: 200 * MeV,
	Cu245: 200 * MeV,
	Cu247: 200 * MeV,
	Cf251: 200 * MeV,
	Th232: 200 * MeV,
	U234:  200 * MeV,
	U238:  200 * MeV,
	Pu238: 200 * MeV,
	Pu240: 200 * MeV,
}
