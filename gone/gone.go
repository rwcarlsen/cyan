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

func Id(nuc string) int {
	cs := C.CString(nuc)
	defer C.free(unsafe.Pointer(cs))
	return int(C.id_str(cs))
}

func IdFromInt(nuc int) int {
	return int(C.id_int(C.int(nuc)))
}

func DecayConst(nuc int) float64 {
	return float64(C.decay_const(C.int(nuc)))
}
