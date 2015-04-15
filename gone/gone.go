package gone

/*
#include <stdlib.h>
#include "cpyne.h"
#cgo LDFLAGS: -lhdf5
*/
import "C"
import (
	"os"
	"unsafe"
)

var NucDataPath = ""

type Nuc int

func (n Nuc) Name() string {
	cname := C.name(C.int(n))
	name := C.GoString(cname)
	C.free(unsafe.Pointer(cname))
	return name
}

func init() {
	fpath := os.Getenv("NUCDATA_PATH")
	if fpath == "" {
		fpath = "./nucdata.h5"
	}
	cs := C.CString(fpath)
	defer C.free(unsafe.Pointer(cs))
	C.init_nuc_data(cs)
}

func SetNucDataPath(fpath string) {
	cs := C.CString(fpath)
	defer C.free(unsafe.Pointer(cs))
	C.init_nuc_data(cs)
}

func Id(nuc string) Nuc {
	cs := C.CString(nuc)
	defer C.free(unsafe.Pointer(cs))
	return Nuc(C.id_str(cs))
}

func IdFromInt(nuc int) Nuc {
	return Nuc(C.id_int(C.int(nuc)))
}

func DecayConst(nuc Nuc) float64 {
	return float64(C.decay_const(C.int(nuc)))
}

type Energy int

const (
	Thermal Energy = iota
	FastFission
	Fast14MeV
)

func FissProdYield(from, to Nuc, e Energy) float64 {
	source := 0
	switch e {
	case Thermal:
		source = 1 // thermal NDS
	case FastFission:
		source = 2 // fast NDS
	case Fast14MeV:
		source = 3 // 14 MeV NDS
	default:
		panic("unsupported fission product yield source")
	}
	return float64(C.fpyield(C.int(from), C.int(to), C.int(source))) / 100
}
