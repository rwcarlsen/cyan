#include "pyne.h"
#include "cpyne.h"
#include <string>
#include <cstring>

int
id_str(const char* nuc) {
  int id = 0;
  try {
    id = pyne::nucname::id(nuc);
  } catch(...) {
    id = -1;
  }
  return id;
}

int
id_int(int nuc) {
  int id = 0;
  try {
    id = pyne::nucname::id(nuc);
  } catch(...) {
    id = -1;
  }
  return id;
}

char*
name(int nuc) {
  std::string s = pyne::nucname::name(nuc);
  char* cstr = new char [s.length()+1];
  std::strcpy(cstr, s.c_str());
  return cstr;
}

double
decay_const(int nuc) {
  return pyne::decay_const(nuc);
}

void
init_nuc_data(const char* fpath) {
  pyne::NUC_DATA_PATH = fpath;
}

double
fpyield(int fromnuc, int tonuc, int source) {
  return pyne::fpyield(fromnuc, tonuc, source, false);
}
