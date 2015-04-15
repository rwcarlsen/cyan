#ifdef __cplusplus
extern "C" {
#endif

  int
  id_str(const char* nuc);

  int
  id_int(int nuc);

  char*
  name(int nuc);

  double
  decay_const(int nuc);

  void
  init_nuc_data(const char* fpath);

  double
  fpyield(int fromnuc, int tonuc, int source);

#ifdef __cplusplus
}
#endif
