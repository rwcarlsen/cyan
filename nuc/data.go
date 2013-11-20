package nuc

// Fissile isltopes
const (
	U235  = 92235
	U233  = 92233
	Pu239 = 94239
	Pu241 = 94241
	Cu243 = 96243
	Cu245 = 96245
	Cu247 = 96247
	Cf251 = 98251
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
	Th232 = 90232
	U234  = 92234
	U238  = 92238
	Pu238 = 94238
	Pu240 = 94240
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
