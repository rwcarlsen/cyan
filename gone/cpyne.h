#ifdef __cplusplus
extern "C" {
#endif

  int
  id_str(const char* nuc);

  int
  id_int(int nuc);

  double
  decay_const(int nuc);

  void
  init_nuc_data(const char* fpath);

#ifdef __cplusplus
}
#endif
