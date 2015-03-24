#include "pyne.h"
#include "cpyne.h"

int
id_str(const char* nuc) {
  return pyne::nucname::id(nuc);
}

int
id_int(int nuc) {
  return pyne::nucname::id(nuc);
}

double
decay_const(int nuc) {
  return pyne::decay_const(nuc);
}

void
init_nuc_data(const char* fpath) {
  pyne::NUC_DATA_PATH = fpath;
}
