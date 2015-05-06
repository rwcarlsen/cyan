#include "nucname.h"
#include <cmath>

// String Transformations
std::string pyne::to_str(int t) {
  std::stringstream ss;
  ss << t;
  return ss.str();
}

std::string pyne::to_str(unsigned int t) {
  std::stringstream ss;
  ss << t;
  return ss.str();
}

std::string pyne::to_str(double t) {
  std::stringstream ss;
  ss << t;
  return ss.str();
}

std::string pyne::to_str(bool t) {
  std::stringstream ss;
  ss << t;
  return ss.str();
}


int pyne::to_int(std::string s) {
  return atoi( s.c_str() );
}

double pyne::to_dbl(std::string s) {
  return strtod( s.c_str(), NULL );
}

double pyne::endftod_cpp(char * s) {
  // Converts string from ENDF only handles "E-less" format but is 5x faster
  int pos, mant, exp;
  double v, dbl_exp;

  mant = exp = 0;
  if (s[2] == '.') {
    // Convert an ENDF float
    if (s[9] == '+' or s[9] == '-') {
      // All these factors of ten are from place values.
      mant = s[8] + 10 * s[7] + 100 * s[6] + 1000 * s[5] + 10000 * s[4] + \
             100000 * s[3] + 1000000 * s[1] - 1111111 * '0';
      exp = s[10] - '0';
      // Make the right power of 10.
      dbl_exp = exp & 01? 10.: 1;
      dbl_exp *= (exp >>= 1) & 01? 100.: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e4: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e8: 1;
      // Adjust for powers of ten from treating mantissa as an integer.
      dbl_exp = (s[9] == '-'? 1/dbl_exp: dbl_exp) * 1.0e-6;
      // Get mantissa sign, apply exponent.
      v = mant * (s[0] == '-'? -1: 1) * dbl_exp;
    }
    else {
      mant = s[7] + 10 * s[6] + 100 * s[5] + 1000 * s[4] + 10000 * s[3] + \
             100000 * s[1] - 111111 * '0';
      exp = s[10] + 10 * s[9] - 11 * '0';
      dbl_exp = exp & 01? 10.: 1;
      dbl_exp *= (exp >>= 1) & 01? 100.: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e4: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e8: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e16: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e32: 1;
      dbl_exp *= (exp >>= 1) & 01? 1.0e64: 1;
      dbl_exp = (s[8] == '-'? 1/dbl_exp: dbl_exp) * 1.0e-5;
      v = mant * (s[0] == '-'? -1: 1) * dbl_exp;
    }
  }

  // Convert an ENDF int to float; we start from the last char in the field and
  // move forward until we hit a non-digit.
  else {
    v = 0;
    mant = 1; // Here we use mant for the place value about to be read in.
    pos = 10;
    while (s[pos] != '-' and s[pos] != '+' and s[pos] != ' ' and pos > 0) {
      v += mant * (s[pos] - '0');
      mant *= 10;
      pos--;
    }
    v *= (s[pos] == '-'? -1: 1);
  }
  return v;
}

double pyne::endftod_f(char * s) {
  return endftod_cpp(s);
}

double (*pyne::endftod)(char * s) = &pyne::endftod_f;

void pyne::use_fast_endftod() {
  pyne::endftod = &pyne::endftod_cpp;
}

std::string pyne::to_upper(std::string s) {
  // change each element of the string to upper case.
  for(unsigned int i = 0; i < s.length(); i++)
    s[i] = toupper(s[i]);
  return s;
}

std::string pyne::to_lower(std::string s) {
  // change each element of the string to lower case
  for(unsigned int i = 0; i < s.length(); i++)
    s[i] = tolower(s[i]);
  return s;
}


std::string pyne::capitalize(std::string s) {
  unsigned int slen = s.length();
  if (slen == 0)
    return s;
  // uppercase the first character
  s[0] = toupper(s[0]);
  // change each subsequent element of the string to lower case
  for(unsigned int i = 1; i < slen; i++)
    s[i] = tolower(s[i]);
  return s;
}


std::string pyne::get_flag(char line[], int max_l) {
  char tempflag [10];
  for (int i = 0; i < max_l; i++)
  {
    if (line[i] == '\t' || line[i] == '\n' || line[i] == ' ' || line[i] == '\0')
    {
      tempflag[i] = '\0';
      break;
    }
    else
      tempflag[i] = line[i];
  }
  return std::string (tempflag);
}



std::string pyne::remove_substring(std::string s, std::string substr) {
  // Removes a substring from the string s
  int n_found = s.find(substr);
  while ( 0 <= n_found ) {
    s.erase( n_found , substr.length() );
    n_found = s.find(substr);
  }
  return s;
}


std::string pyne::remove_characters(std::string s, std::string chars) {
  // Removes all characters in the string chars from the string s
  for (int i = 0; i < chars.length(); i++ ) {
    s = remove_substring(s, chars.substr(i, 1) );
  }
  return s;
}


std::string pyne::replace_all_substrings(std::string s, std::string substr, std::string repstr) {
  // Replaces all instance of substr in s with the string repstr
  int n_found = s.find(substr);
  while ( 0 <= n_found ) {
    s.replace( n_found , substr.length(), repstr );
    n_found = s.find(substr);
  }
  return s;
};



std::string pyne::last_char(std::string s) {
    // Returns the last character in a string.
    return s.substr(s.length()-1, 1);
}


std::string pyne::slice_from_end(std::string s, int n, int l) {
  // Returns the slice of a string using negative indices.
  return s.substr(s.length()+n, l);
}


bool pyne::ternary_ge(int a, int b, int c) {
  // Returns true id a <= b <= c and flase otherwise.
  return (a <= b && b <= c);
}


bool pyne::contains_substring(std::string s, std::string substr) {
  // Returns a boolean based on if the sub is in s.
  int n = s.find(substr);
  return ( 0 <= n && n < s.length() );
}


std::string pyne::natural_naming(std::string name) {
  // Calculates a version on the string name that is a valid
  // variable name, ie it uses only word characters.
  std::string nat_name (name);

  // Replace Whitespace characters with underscores
  nat_name = pyne::replace_all_substrings(nat_name, " ",  "_");
  nat_name = pyne::replace_all_substrings(nat_name, "\t", "_");
  nat_name = pyne::replace_all_substrings(nat_name, "\n", "_");

  // Remove non-word characters
  int n = 0;
  while ( n < nat_name.length() ) {
    if ( pyne::words.find(nat_name[n]) == std::string::npos )
      nat_name.erase(n, 1);
    else
      n++;
  }

  // Make sure that the name in non-empty before continuing
  if (nat_name.length() == 0)
    return nat_name;

  // Make sure that the name doesn't begin with a number.
  if ( pyne::digits.find(nat_name[0]) != std::string::npos)
    nat_name.insert(0, "_");

  return nat_name;
};


//
// Math Helpers
//

double pyne::slope(double x2, double y2, double x1, double y1) {
  // Finds the slope of a line.
  return (y2 - y1) / (x2 - x1);
};


double pyne::solve_line(double x, double x2, double y2, double x1, double y1) {
  return (slope(x2,y2,x1,y1) * (x - x2)) + y2;
};


double pyne::tanh(double x) {
  return std::tanh(x);
};

double pyne::coth(double x) {
  return 1.0 / std::tanh(x);
};



// Message Helpers
 
bool pyne::USE_WARNINGS = true;

bool pyne::toggle_warnings(){
  USE_WARNINGS = !USE_WARNINGS;
  return USE_WARNINGS;
}

void pyne::warning(std::string s){
  // Prints a warning message
  if (USE_WARNINGS){
    std::cout << "\033[1;33m WARNING: \033[0m" << s << "\n"; 
  }  
}

namespace pyne {

namespace nucname {
#define TOTAL_STATE_MAPS 922
std::map<int, int> state_id_map;
int map_nuc_ids [TOTAL_STATE_MAPS] = {
110240001,
130240001,
130260001,
130320002,
170340001,
170380001,
190380001,
190380015,
210420002,
210430001,
210440004,
230440001,
210450001,
210460002,
230460001,
210500001,
250500001,
250520001,
260520041,
260530022,
270540001,
210560001,
210560004,
250580001,
270580001,
270580002,
230600000,
230600001,
250600001,
270600001,
250620001,
270620001,
230640001,
250640002,
260650003,
260670002,
290670023,
280690001,
280690008,
300690001,
340690004,
290700001,
290700003,
350700006,
280710002,
300710001,
320710002,
300730001,
300730002,
320730002,
340730001,
360730004,
310740002,
350740002,
290750001,
290750002,
300750001,
320750002,
330750004,
280760004,
290760001,
350760002,
300770002,
320770001,
330770004,
340770001,
350770001,
300780004,
310780004,
350780004,
370780003,
390780001,
320790001,
330790007,
340790001,
350790001,
360790001,
310800001,
350800002,
390800001,
390800003,
320810001,
340810001,
360810002,
370810001,
330820001,
340820015,
350820001,
370820001,
410820003,
340830001,
360830002,
380830002,
390830001,
310840001,
350840001,
360840019,
360840061,
370840002,
390840002,
410840007,
360850001,
370850003,
380850002,
390850001,
400850002,
410850003,
410850005,
370860002,
390860002,
410860001,
410860002,
380870001,
390870001,
400870002,
350880003,
410880001,
430880000,
430880001,
390890001,
400890001,
410890001,
420890002,
430890001,
370900001,
390900002,
400900003,
410900002,
410900007,
430900001,
430900006,
390910001,
400910040,
410910001,
420910001,
430910001,
440910001,
450910001,
410920001,
450920001,
390930002,
410930001,
420930016,
430930001,
440930001,
470940001,
470940002,
390970001,
390970029,
410970001,
430970001,
450970001,
370980001,
390980005,
410980001,
450980001,
410990001,
430990002,
450990001,
470990002,
371000001,
391000004,
411000001,
411000009,
411000012,
431000002,
431000004,
451000004,
471000001,
471010002,
411020001,
431020001,
451020005,
471020001,
441030005,
451030001,
471030002,
491030001,
411040004,
451040003,
471040001,
491040003,
451050001,
471050001,
491050001,
451060001,
471060001,
491060001,
431070000,
461070002,
471070001,
491070001,
401080003,
461090002,
471090001,
491090001,
491090021,
451100000,
451100001,
471100002,
491100001,
461110002,
471110001,
491110001,
451120000,
451120001,
491120001,
491120004,
491120010,
471130001,
481130001,
491130001,
501130001,
451140005,
491140001,
491140005,
531140005,
461150001,
471150001,
481150001,
491150001,
521150001,
451160001,
471160001,
471160004,
511160003,
551160001,
471180004,
491180001,
491180003,
511180007,
531180002,
551180001,
471190000,
471190001,
481190002,
491190001,
501190002,
511190072,
521190002,
551190001,
451200002,
471200002,
491200001,
491200002,
511200001,
531200013,
551200001,
571200000,
461210001,
481210002,
491210001,
501210001,
521210002,
551210001,
451220002,
471220001,
471220002,
491220001,
491220005,
511220005,
511220006,
551220007,
551220008,
481230003,
491230001,
501230001,
521230002,
551230005,
461240004,
491240002,
501240016,
511240001,
511240002,
551240025,
481250001,
491250001,
501250001,
521250002,
541250002,
571250005,
461260003,
461260004,
491260001,
511260001,
511260002,
481270006,
491270001,
491270009,
501270001,
521270002,
541270002,
561270002,
571270001,
581270001,
461280004,
491280003,
501280003,
511280001,
571280001,
471290001,
481290001,
491290001,
491290010,
491290012,
491290013,
501290001,
501290017,
501290018,
501290025,
511290011,
511290012,
511290023,
521290001,
541290002,
551290010,
561290001,
571290002,
601290001,
601290003,
491300001,
491300002,
491300003,
501300002,
511300001,
531300001,
551300004,
561300030,
591300002,
491310001,
491310004,
501310001,
521310001,
521310033,
541310002,
561310002,
571310006,
581310001,
591310002,
501320006,
511320001,
521320006,
521320022,
531320003,
541320030,
571320004,
581320030,
491330001,
521330002,
531330016,
531330059,
531330065,
541330001,
561330002,
581330001,
591330003,
601330001,
611330005,
621330000,
511340002,
521340003,
531340005,
541340007,
601340017,
611340000,
611340001,
521350010,
541350002,
551350010,
561350002,
581350004,
591350004,
601350001,
611350000,
611350003,
501360003,
531360006,
551360001,
561360005,
611360000,
611360001,
631360001,
561370002,
581370002,
601370004,
501380003,
551380003,
581380005,
591380005,
581390002,
601390002,
611390001,
621390004,
641390001,
591400003,
591400015,
601400009,
611400008,
631400004,
601410002,
621410002,
631410001,
641410004,
651410001,
591420001,
591420024,
601420004,
611420012,
631420031,
641420019,
641420020,
651420003,
621430002,
621430043,
641430002,
651430001,
661430003,
551440004,
591440001,
651440004,
651440006,
651440007,
671440003,
641450002,
651450004,
661450002,
681450002,
571460001,
631460013,
651460022,
651460026,
661460008,
651470001,
661470002,
681470002,
691470001,
591480000,
591480001,
611480003,
651480001,
671480001,
671480012,
681480008,
651490001,
661490027,
671490001,
681490002,
631500001,
651500002,
671500001,
691500005,
581510001,
621510012,
631510002,
651510003,
671510001,
681510021,
691510001,
691510012,
701510001,
701510005,
701510010,
611520004,
611520014,
631520001,
631520016,
651520006,
671520001,
691520006,
691520018,
691520019,
701520006,
621530006,
641530003,
641530008,
651530003,
671530001,
691530001,
601540003,
611540000,
611540001,
631540013,
651540001,
651540002,
711540015,
721540006,
641550006,
661550009,
671550002,
691550001,
711550001,
711550004,
611560002,
651560002,
651560004,
671560001,
671560012,
711560001,
721560004,
641570012,
661570005,
651580003,
651580019,
671580001,
671580007,
711580000,
621590006,
641590002,
661590009,
671590003,
671600001,
671600006,
691600002,
711600001,
671610002,
681610014,
691610001,
711610004,
671620003,
691620020,
711620008,
711620009,
751620001,
671630003,
751630001,
671640003,
691640001,
771640001,
661650002,
751650001,
771650001,
671660001,
691660006,
711660001,
711660002,
681670003,
711670001,
751670001,
671680001,
711680013,
771680001,
701690001,
711690001,
751690001,
771690001,
671700001,
711700008,
771700001,
711710001,
721710001,
771710001,
781710002,
711720001,
711720005,
751720001,
771720002,
791720001,
771730000,
771730029,
791730001,
711740003,
771740001,
701750007,
711750053,
791750001,
701760005,
711760001,
731760012,
731760090,
791760001,
791760002,
691770000,
701770006,
711770029,
711770203,
721770048,
721770107,
791770002,
711780003,
721780005,
721780109,
731780000,
731780059,
731780094,
731780139,
711790006,
721790005,
721790046,
731790117,
741790002,
751790137,
791790007,
811790001,
711800010,
721800007,
731800002,
721810025,
721810078,
761810001,
811810002,
721820009,
721820026,
731820001,
731820029,
751820001,
761820029,
741830007,
751830058,
761830002,
781830001,
811830002,
721840005,
751840005,
771840007,
781840034,
791840003,
741850006,
781850002,
791850001,
801850004,
811850003,
751860004,
771860001,
811860000,
811860005,
831860001,
791870002,
801870001,
811870002,
821870001,
831870002,
751880007,
811880001,
761890001,
771890006,
771890084,
791890003,
801890002,
811890001,
821890001,
831890002,
831890003,
731900002,
741900006,
751900003,
761900032,
771900002,
771900037,
791900014,
811900000,
811900001,
811900006,
831900000,
831900001,
761910001,
771910003,
771910071,
791910004,
801910035,
811910002,
821910002,
831910002,
751920002,
751920003,
761920047,
761920112,
771920003,
771920015,
791920004,
791920015,
811920002,
811920008,
821920011,
821920014,
821920017,
821920020,
821920021,
831920001,
841920006,
851920000,
851920001,
771930002,
781930005,
791930004,
801930003,
811930002,
821930001,
831930001,
851930001,
851930002,
751940001,
751940002,
751940003,
771940007,
771940012,
791940003,
791940008,
811940001,
831940001,
831940002,
851940000,
851940001,
761950002,
761950004,
771950002,
781950007,
791950004,
791950055,
801950003,
811950002,
821950002,
831950001,
841950002,
851950001,
861950001,
751960001,
771960004,
791960003,
791960054,
811960006,
831960002,
831960003,
841960015,
761970001,
771970002,
781970009,
791970004,
801970004,
811970002,
821970002,
831970001,
841970002,
851970001,
861970001,
761980006,
761980010,
771980001,
791980050,
811980007,
811980012,
831980001,
831980003,
851980001,
871980001,
781990008,
791990006,
801990007,
811990003,
821990003,
831990001,
841990002,
861990001,
812000010,
832000001,
832000003,
852000001,
852000003,
802010013,
812010003,
822010004,
832010001,
842010003,
862010001,
872010001,
882010000,
782020003,
822020014,
852020001,
852020002,
872020001,
822030006,
822030053,
832030006,
842030005,
862030001,
882030001,
812040029,
822040021,
832040008,
832040038,
852040001,
872040001,
872040002,
802050008,
822050009,
842050010,
842050017,
882050001,
812060045,
832060016,
872060001,
872060002,
892060001,
812070002,
822070003,
832070036,
842070014,
862070007,
882070001,
802080004,
832080018,
802100002,
802100005,
832100002,
822110014,
832110021,
842110015,
852110076,
872110013,
872110019,
832120005,
832120012,
842120030,
852120004,
882130005,
852140006,
862140004,
862140005,
872140001,
902140004,
832150009,
862150013,
902150003,
872160001,
832170005,
892170010,
902170001,
912170001,
872180002,
922180001,
892220001,
912340002,
922350001,
932360001,
952360001,
942370003,
922380101,
932380128,
942380041,
942380044,
952380001,
942390090,
942390094,
952390011,
932400001,
942400102,
952400057,
962400002,
962400003,
942410106,
942410107,
952410075,
962410007,
932420007,
942420044,
942420045,
952420002,
952420141,
962420004,
962420005,
972420002,
972420003,
942440032,
952440001,
952440112,
952440113,
962440009,
962440013,
962440014,
972440004,
982440002,
942450024,
952450021,
962450061,
972450003,
1012450001,
952460001,
952460008,
972460000,
982460002,
992460000,
1012460000,
1012460001,
1002470001,
1002470002,
972480001,
992500001,
1002500001,
1002500002,
1022500001,
1022510002,
1002530008,
1022530003,
1022530030,
1022530031,
1022530032,
1032530000,
1032530001,
992540002,
1012540000,
1012540001,
1022540011,
1032550001,
1032550027,
992560001,
1002560022,
1042560007,
1042560009,
1042560012,
1042570002,
1052570002,
1012580001,
1052580001,
1042610001,
1072620001,
1062630003,
1062650001,
1082650001,
1082670002,
1102700001,
1102710001,
1082770001,
};
int map_metastable [TOTAL_STATE_MAPS] = {1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
2,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
3,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
3,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
2,
1,
2,
3,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
3,
4,
1,
2,
3,
4,
1,
2,
3,
1,
1,
1,
1,
1,
1,
2,
1,
2,
3,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
2,
3,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
3,
1,
2,
1,
2,
1,
1,
1,
2,
3,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
2,
1,
1,
2,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
2,
1,
1,
1,
2,
1,
2,
3,
4,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
3,
1,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
2,
1,
1,
2,
3,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
2,
1,
2,
1,
2,
1,
2,
3,
4,
5,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
3,
1,
2,
1,
2,
1,
1,
2,
1,
2,
1,
2,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
2,
1,
2,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
2,
1,
1,
2,
1,
2,
1,
2,
1,
2,
1,
2,
1,
1,
2,
3,
1,
2,
3,
1,
1,
1,
1,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
2,
1,
1,
1,
1,
2,
1,
1,
1,
1,
2,
3,
4,
1,
1,
1,
1,
2,
1,
1,
2,
1,
1,
1,
2,
3,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
1,
};
} // namespace nucname
} // namespace pyne


/*** Constructs the LL to zz Dictionary ***/
pyne::nucname::name_zz_t pyne::nucname::get_name_zz() {
  pyne::nucname::name_zz_t lzd;

  lzd["Be"] = 04;
  lzd["Ba"] = 56;
  lzd["Bh"] = 107;
  lzd["Bi"] = 83;
  lzd["Bk"] = 97;
  lzd["Br"] = 35;
  lzd["Ru"] = 44;
  lzd["Re"] = 75;
  lzd["Rf"] = 104;
  lzd["Rg"] = 111;
  lzd["Ra"] = 88;
  lzd["Rb"] = 37;
  lzd["Rn"] = 86;
  lzd["Rh"] = 45;
  lzd["Tm"] = 69;
  lzd["H"] = 01;
  lzd["P"] = 15;
  lzd["Ge"] = 32;
  lzd["Gd"] = 64;
  lzd["Ga"] = 31;
  lzd["Os"] = 76;
  lzd["Hs"] = 108;
  lzd["Zn"] = 30;
  lzd["Ho"] = 67;
  lzd["Hf"] = 72;
  lzd["Hg"] = 80;
  lzd["He"] = 02;
  lzd["Pr"] = 59;
  lzd["Pt"] = 78;
  lzd["Pu"] = 94;
  lzd["Pb"] = 82;
  lzd["Pa"] = 91;
  lzd["Pd"] = 46;
  lzd["Po"] = 84;
  lzd["Pm"] = 61;
  lzd["C"] = 6;
  lzd["K"] = 19;
  lzd["O"] = 8;
  lzd["S"] = 16;
  lzd["W"] = 74;
  lzd["Eu"] = 63;
  lzd["Es"] = 99;
  lzd["Er"] = 68;
  lzd["Md"] = 101;
  lzd["Mg"] = 12;
  lzd["Mo"] = 42;
  lzd["Mn"] = 25;
  lzd["Mt"] = 109;
  lzd["U"] = 92;
  lzd["Fr"] = 87;
  lzd["Fe"] = 26;
  lzd["Fm"] = 100;
  lzd["Ni"] = 28;
  lzd["No"] = 102;
  lzd["Na"] = 11;
  lzd["Nb"] = 41;
  lzd["Nd"] = 60;
  lzd["Ne"] = 10;
  lzd["Zr"] = 40;
  lzd["Np"] = 93;
  lzd["B"] = 05;
  lzd["Co"] = 27;
  lzd["Cm"] = 96;
  lzd["F"] = 9;
  lzd["Ca"] = 20;
  lzd["Cf"] = 98;
  lzd["Ce"] = 58;
  lzd["Cd"] = 48;
  lzd["V"] = 23;
  lzd["Cs"] = 55;
  lzd["Cr"] = 24;
  lzd["Cu"] = 29;
  lzd["Sr"] = 38;
  lzd["Kr"] = 36;
  lzd["Si"] = 14;
  lzd["Sn"] = 50;
  lzd["Sm"] = 62;
  lzd["Sc"] = 21;
  lzd["Sb"] = 51;
  lzd["Sg"] = 106;
  lzd["Se"] = 34;
  lzd["Yb"] = 70;
  lzd["Db"] = 105;
  lzd["Dy"] = 66;
  lzd["Ds"] = 110;
  lzd["La"] = 57;
  lzd["Cl"] = 17;
  lzd["Li"] = 03;
  lzd["Tl"] = 81;
  lzd["Lu"] = 71;
  lzd["Lr"] = 103;
  lzd["Th"] = 90;
  lzd["Ti"] = 22;
  lzd["Te"] = 52;
  lzd["Tb"] = 65;
  lzd["Tc"] = 43;
  lzd["Ta"] = 73;
  lzd["Ac"] = 89;
  lzd["Ag"] = 47;
  lzd["I"] = 53;
  lzd["Ir"] = 77;
  lzd["Am"] = 95;
  lzd["Al"] = 13;
  lzd["As"] = 33;
  lzd["Ar"] = 18;
  lzd["Au"] = 79;
  lzd["At"] = 85;
  lzd["In"] = 49;
  lzd["Y"] = 39;
  lzd["N"] = 07;
  lzd["Xe"] = 54;
  lzd["Cn"] = 112;
  lzd["Fl"] = 114;
  lzd["Lv"] = 116;

  return lzd;
};
pyne::nucname::name_zz_t pyne::nucname::name_zz = pyne::nucname::get_name_zz();


/*** Constructs zz to LL dictionary **/
pyne::nucname::zzname_t pyne::nucname::get_zz_name()
{
  zzname_t zld;
  for (name_zz_iter i = name_zz.begin(); i != name_zz.end(); i++)
  {
    zld[i->second] = i->first;
  }
  return zld;
};
pyne::nucname::zzname_t pyne::nucname::zz_name = pyne::nucname::get_zz_name();



/*** Constructs the fluka to zz Dictionary ***/
pyne::nucname::name_zz_t pyne::nucname::get_fluka_zz() {
  pyne::nucname::name_zz_t fzd;

  fzd["BERYLLIU"] = 40000000;
  fzd["BARIUM"]   = 560000000;
  fzd["BOHRIUM"]  = 1070000000;   // No fluka
  fzd["BISMUTH"]  = 830000000;
  fzd["BERKELIU"] = 970000000;    // No fluka 
  fzd["BROMINE"]  = 350000000;
  fzd["RUTHENIU"] = 440000000;    // No fluka
  fzd["RHENIUM"]  = 750000000;
  fzd["RUTHERFO"] = 1040000000;   
  fzd["ROENTGEN"] = 1110000000;
  fzd["RADIUM"]   = 880000000;    // No fluka
  fzd["RUBIDIUM"] = 370000000;    // No fluka
  fzd["RADON"]    = 860000000;    // no fluka
  fzd["RHODIUM"]  = 450000000;    // no fluka
  fzd["THULIUM"]  = 690000000;    // no fluka
  fzd["HYDROGEN"] = 10000000;        
  fzd["PHOSPHO"]  = 150000000;
  fzd["GERMANIU"] = 320000000;
  fzd["GADOLINI"] = 640000000;
  fzd["GALLIUM"]  = 310000000;
  fzd["OSMIUM"]   = 760000000;    // no fluka
  fzd["HASSIUM"]  = 1080000000;
  fzd["ZINC"]     = 300000000;
  fzd["HOLMIUM"]  = 670000000;    // no fluka
  fzd["HAFNIUM"]  = 720000000;
  fzd["MERCURY"]  = 800000000;
  fzd["HELIUM"]   = 20000000;
  fzd["PRASEODY"] = 590000000;   // no fluka
  fzd["PLATINUM"] = 780000000;
  fzd["239-PU"]   = 940000000;   // "239-PU"
  fzd["LEAD"]     = 820000000;
  fzd["PROTACTI"] = 910000000;   // no fluka
  fzd["PALLADIU"] = 460000000;   // no fluka
  fzd["POLONIUM"] = 840000000;   // no fluka 
  fzd["PROMETHI"] = 610000000;   // no fluka
  fzd["CARBON"]   = 60000000;
  fzd["POTASSIU"] = 190000000;
  fzd["OXYGEN"]   = 80000000;
  fzd["SULFUR"]   = 160000000;
  fzd["TUNGSTEN"] = 740000000;
  fzd["EUROPIUM"] = 630000000;
  fzd["EINSTEIN"] = 990000000;   // no fluka
  fzd["ERBIUM"]   = 680000000;   // no fluka
  fzd["MENDELEV"] = 1010000000;  // no fluka
  fzd["MAGNESIU"] = 120000000;
  fzd["MOLYBDEN"] = 420000000;
  fzd["MANGANES"] = 250000000;
  fzd["MEITNERI"] = 1090000000;  // no fluka
  fzd["URANIUM"]  = 920000000;
  fzd["FRANCIUM"] = 870000000;   // no fluka
  fzd["IRON"]     = 260000000;
  fzd["FERMIUM"]  = 1000000000;  // no fluka
  fzd["NICKEL"]   = 280000000;
  fzd["NITROGEN"] = 70000000;
  fzd["NOBELIUM"] = 1020000000;  // no fluka
  fzd["SODIUM"]   = 110000000;
  fzd["NIOBIUM"]  = 410000000;
  fzd["NEODYMIU"] = 600000000;
  fzd["NEON"]     = 100000000;
  fzd["ZIRCONIU"] = 400000000;
  fzd["NEPTUNIU"] = 930000000;   // no fluka
  fzd["BORON"]    = 50000000;
  fzd["COBALT"]   = 270000000;
  fzd["CURIUM"]   = 960000000;   // no fluka
  fzd["FLUORINE"] = 90000000;
  fzd["CALCIUM"]  = 200000000;
  fzd["CALIFORN"] = 980000000;   // no fluka
  fzd["CERIUM"]   = 580000000;
  fzd["CADMIUM"]  = 480000000;
  fzd["VANADIUM"] = 230000000;
  fzd["CESIUM"]   = 550000000;
  fzd["CHROMIUM"] = 240000000;
  fzd["COPPER"]   = 290000000;
  fzd["STRONTIU"] = 380000000;
  fzd["KRYPTON"]  = 360000000;
  fzd["SILICON"]  = 140000000;
  fzd["TIN"]      = 500000000;
  fzd["SAMARIUM"] = 620000000;
  fzd["SCANDIUM"] = 210000000;
  fzd["ANTIMONY"] = 510000000;
  fzd["SEABORGI"] = 1060000000;  // no fluka
  fzd["SELENIUM"] = 340000000;   // no fluka
  fzd["YTTERBIU"] = 700000000;   // no fluka
  fzd["DUBNIUM"]  = 1050000000;  // no fluka
  fzd["DYSPROSI"] = 660000000;   // no fluka
  fzd["DARMSTAD"] = 1100000000;  // no fluka
  fzd["LANTHANU"] = 570000000;
  fzd["CHLORINE"] = 170000000;
  fzd["LITHIUM"]  = 030000000;
  fzd["THALLIUM"] = 810000000;   // no fluka
  fzd["LUTETIUM"] = 710000000;   // no fluka
  fzd["LAWRENCI"] = 1030000000;  // no fluka
  fzd["THORIUM"]  = 900000000;   // no fluka
  fzd["TITANIUM"] = 220000000;
  fzd["TELLURIU"] = 520000000;   // no fluka
  fzd["TERBIUM"]  = 650000000;
  fzd["99-TC"]    = 430000000;   // "99-TC"
  fzd["TANTALUM"] = 730000000;
  fzd["ACTINIUM"] = 890000000;   // no fluka
  fzd["SILVER"]   = 470000000;
  fzd["IODINE"]   = 530000000;
  fzd["IRIDIUM"]  = 770000000;
  fzd["241-AM"]   = 950000000;   // "241-AM"
  fzd["ALUMINUM"] = 130000000;
  fzd["ARSENIC"]  = 330000000;
  fzd["ARGON"]    = 180000000;
  fzd["GOLD"]     = 790000000;
  fzd["ASTATINE"] = 850000000;   // no fluka
  fzd["INDIUM"]   = 490000000;
  fzd["YTTRIUM"]  = 390000000;
  fzd["XENON"]    = 540000000;
  fzd["COPERNIC"] = 1120000000;  // no fluka
  fzd["UNUNQUAD"] = 1140000000;  // no fluka:  UNUNQUADIUM,  "Flerovium"
  fzd["UNUNHEXI"] = 1160000000;  // no fluka:  UNUNHEXIUM , "Livermorium" 
  fzd["HYDROG-1"] = 10010000;        
  fzd["DEUTERIU"] = 10020000;        
  fzd["TRITIUM"]  = 10040000;        
  fzd["HELIUM-3"] = 20030000;
  fzd["HELIUM-4"] = 20040000;
  fzd["LITHIU-6"] = 30060000;
  fzd["LITHIU-7"] = 30070000;
  fzd["BORON-10"] = 50100000;
  fzd["BORON-11"] = 50110000;
  fzd["90-SR"]    = 380900000;   // fluka "90-SR"
  fzd["129-I"]    = 531290000;   // fluka "129-I"
  fzd["124-XE"]   = 541240000;   // fluka "124-XE"
  fzd["126-XE"]   = 541260000;   // fluka "126-XE"
  fzd["128-XE"]   = 541280000;   // fluka "128-XE"
  fzd["130-XE"]   = 541300000;   // fluka "130-XE"
  fzd["131-XE"]   = 541310000;   // fluka "131-XE"
  fzd["132-XE"]   = 541320000;   // fluka "132-XE"
  fzd["134-XE"]   = 541340000;   // fluka "134-XE"
  fzd["135-XE"]   = 541350000;   // fluka "135-XE"
  fzd["136-XE"]   = 541360000;   // fluka "136-XE"
  fzd["135-CS"]   = 551350000;   // fluka "135-CS"
  fzd["137-CS"]   = 551370000;   // fluka "137-CS"
  fzd["230-TH"]   = 902300000;   // fluka "230-TH"
  fzd["232-TH"]   = 902320000;   // fluka "232-TH"
  fzd["233-U"]    = 922330000;   // fluka "233-U"
  fzd["234-U"]    = 922340000;   // fluka "234-U"
  fzd["235-U"]    = 922350000;   // fluka "235-U"
  fzd["238-U"]    = 922380000;   // fluka "238-U"

  return fzd;
};
pyne::nucname::name_zz_t pyne::nucname::fluka_zz = pyne::nucname::get_fluka_zz();


/*** Constructs zz to fluka dictionary **/
pyne::nucname::zzname_t pyne::nucname::get_zz_fluka()
{
  zzname_t zfd;
  for (name_zz_iter i = fluka_zz.begin(); i != fluka_zz.end(); i++)
  {
    zfd[i->second] = i->first;
  }
  return zfd;
};
pyne::nucname::zzname_t pyne::nucname::zz_fluka = pyne::nucname::get_zz_fluka();



/******************************************/
/*** Define useful elemental group sets ***/
/******************************************/

pyne::nucname::zz_group pyne::nucname::name_to_zz_group(pyne::nucname::name_group eg)
{
  zz_group zg;
  for (name_group_iter i = eg.begin(); i != eg.end(); i++)
    zg.insert(name_zz[*i]);
  return zg;
};

// Lanthanides
pyne::nucname::name_t pyne::nucname::LAN_array[15] = {"La", "Ce", "Pr", "Nd", 
  "Pm", "Sm", "Eu", "Gd", "Tb", "Dy", "Ho", "Er", "Tm", "Yb", "Lu"};
pyne::nucname::name_group pyne::nucname::LAN (pyne::nucname::LAN_array, 
                                              pyne::nucname::LAN_array+15);
pyne::nucname::zz_group pyne::nucname::lan = \
  pyne::nucname::name_to_zz_group(pyne::nucname::LAN);

// Actinides
pyne::nucname::name_t pyne::nucname::ACT_array[15] = {"Ac", "Th", "Pa", "U", 
  "Np", "Pu", "Am", "Cm", "Bk", "Cf", "Es", "Fm", "Md", "No", "Lr"};
pyne::nucname::name_group pyne::nucname::ACT (pyne::nucname::ACT_array, pyne::nucname::ACT_array+15);
pyne::nucname::zz_group pyne::nucname::act = pyne::nucname::name_to_zz_group(pyne::nucname::ACT);

// Transuarnics
pyne::nucname::name_t pyne::nucname::TRU_array[22] = {"Np", "Pu", "Am", "Cm", 
  "Bk", "Cf", "Es", "Fm", "Md", "No", "Lr", "Rf", "Db", "Sg", "Bh", "Hs", "Mt", 
  "Ds", "Rg", "Cn", "Fl", "Lv"};
pyne::nucname::name_group pyne::nucname::TRU (pyne::nucname::TRU_array, 
                                              pyne::nucname::TRU_array+22);
pyne::nucname::zz_group pyne::nucname::tru = \
  pyne::nucname::name_to_zz_group(pyne::nucname::TRU);

//Minor Actinides
pyne::nucname::name_t pyne::nucname::MA_array[10] = {"Np", "Am", "Cm", "Bk", 
  "Cf", "Es", "Fm", "Md", "No", "Lr"};
pyne::nucname::name_group pyne::nucname::MA (pyne::nucname::MA_array, 
                                             pyne::nucname::MA_array+10);
pyne::nucname::zz_group pyne::nucname::ma = \
  pyne::nucname::name_to_zz_group(pyne::nucname::MA);

//Fission Products
pyne::nucname::name_t pyne::nucname::FP_array[88] = {"Ag", "Al", "Ar", "As", 
  "At", "Au", "B",  "Ba", "Be", "Bi", "Br", "C",  "Ca", "Cd", "Ce", "Cl", "Co",
  "Cr", "Cs", "Cu", "Dy", "Er", "Eu", "F",  "Fe", "Fr", "Ga", "Gd", "Ge", "H",  
  "He", "Hf", "Hg", "Ho", "I",  "In", "Ir", "K",  "Kr", "La", "Li", "Lu", "Mg", 
  "Mn", "Mo", "N",  "Na", "Nb", "Nd", "Ne", "Ni", "O",  "Os", "P",  "Pb", "Pd", 
  "Pm", "Po", "Pr", "Pt", "Ra", "Rb", "Re", "Rh", "Rn", "Ru", "S",  "Sb", "Sc", 
  "Se", "Si", "Sm", "Sn", "Sr", "Ta", "Tb", "Tc", "Te", "Ti", "Tl", "Tm", "V",  
  "W",  "Xe", "Y",  "Yb", "Zn", "Zr"};
pyne::nucname::name_group pyne::nucname::FP (pyne::nucname::FP_array, 
                                             pyne::nucname::FP_array+88);
pyne::nucname::zz_group pyne::nucname::fp = \
  pyne::nucname::name_to_zz_group(pyne::nucname::FP);


/***************************/
/*** isnuclide functions ***/
/***************************/

bool pyne::nucname::isnuclide(std::string nuc) {
  int n;
  try {
    n = id(nuc);
  }
  catch(NotANuclide) {
    return false;
  }
  catch(IndeterminateNuclideForm) {
    return false;
  };
  return isnuclide(n);
};

bool pyne::nucname::isnuclide(const char * nuc) {
  return isnuclide(std::string(nuc));
};

bool pyne::nucname::isnuclide(int nuc) {
  int n;
  try {
    n = id(nuc);
  }
  catch(NotANuclide) {
    return false;
  }
  catch(IndeterminateNuclideForm) {
    return false;
  };
  if (n <= 10000000)
    return false;
  int zzz = n / 10000000;
  int aaa = (n % 10000000) / 10000;
  if (aaa == 0)
    return false;  // is element
  else if (aaa < zzz)
    return false;
  return true;
};



/********************/
/*** id functions ***/
/********************/
int pyne::nucname::id(int nuc) {
  if (nuc < 0)
    throw NotANuclide(nuc, "");

  int newnuc;
  int zzz = nuc / 10000000;     // ZZZ ?
  int aaassss = nuc % 10000000; // AAA-SSSS ?
  int aaa = aaassss / 10000;    // AAA ?
  int ssss = aaassss % 10000;   // SSSS ?
  // Nuclide must already be in id form
  if (0 < zzz && zzz <= aaa && aaa <= zzz * 7) {
    // Normal nuclide
    if (5 < ssss){
    // Unphysical metastable state warning 
     warning("You have indicated a metastable state of " + pyne::to_str(ssss) + ". Metastable state above 5, possibly unphysical. ");
    }
    return nuc;
  } else if (aaassss == 0 && 0 < zz_name.count(zzz)) {
    // Natural elemental nuclide:  ie for Uranium = 920000000
    return nuc;
  } else if (nuc < 1000 && 0 < zz_name.count(nuc))
    //  Gave Z-number
    return nuc * 10000000;

  // Not in id form, try  ZZZAAAM form.
  zzz = nuc / 10000;     // ZZZ ?
  aaassss = nuc % 10000; // AAA-SSSS ?
  aaa = aaassss / 10;    // AAA ?
  ssss = nuc % 10;       // SSSS ?          
  if (zzz <= aaa && aaa <= zzz * 7) {
    // ZZZAAAM nuclide
    if (5 < ssss){
    // Unphysical metastable state warning
      warning("You have indicated a metastable state of " + pyne::to_str(ssss) + ". Metastable state above 5, possibly unphysical. ");
    }
    return (zzz*10000000) + (aaa*10000) + (nuc%10);
  } else if (aaa <= zzz && zzz <= aaa * 7 && 0 < zz_name.count(aaa)) {
    // Cinder-form (aaazzzm), ie 2350920
    if (5 < ssss){
    // Unphysical metastable state warning
      warning("You have indicated a metastable state of " + pyne::to_str(ssss) + ". Metastable state above 5, possibly unphysical. ");
    }
    return (aaa*10000000) + (zzz*10000) + (nuc%10);
  }
  //else if (aaassss == 0 && 0 == zz_name.count(nuc/1000) && 0 < zz_name.count(zzz))
  else if (aaassss == 0 && 0 < zz_name.count(zzz)) {
    // zzaaam form natural nuclide
    return zzz * 10000000;
  }

  if (nuc >= 1000000){
    // From now we assume no metastable info has been given.
    throw IndeterminateNuclideForm(nuc, "");
  };

  // Nuclide is not in zzaaam form, 
  // Try MCNP form, ie zzaaa
  // This is the same form as SZA for the 0th state.
  zzz = nuc / 1000;
  aaa = nuc % 1000; 
  if (zzz <= aaa) {
    if (aaa - 400 < 0) {
      if (nuc == 95242)
        return nuc * 10000 + 1;  // special case MCNP Am-242m
      else
        return nuc * 10000;  // Nuclide in normal MCNP form
    } else {
      // Nuclide in MCNP metastable form
      if (nuc == 95642)
        return (95642 - 400)*10000;  // special case MCNP Am-242
      nuc = ((nuc - 400) * 10000) + 1;
      while (3.0 < (float ((nuc/10000)%1000) / float (nuc/10000000)))
        nuc -= 999999;
      return nuc;
    }
  } else if (aaa == 0 && 0 < zz_name.count(zzz)) {
    // MCNP form natural nuclide
    return zzz * 10000000;
  }

  // Not a normal nuclide, might be a 
  // Natural elemental nuclide.  
  // ie 92 for Uranium = 920000
  if (0 < zz_name.count(nuc))
    return nuc * 10000000;
  throw IndeterminateNuclideForm(nuc, "");
};

int pyne::nucname::id(const char * nuc) {
  std::string newnuc (nuc);
  return id(newnuc);
};

int pyne::nucname::id(std::string nuc) {
  size_t npos = std::string::npos;
  if (nuc.empty())
    throw NotANuclide(nuc, "<empty>");
  int newnuc;
  std::string elem_name;
  int dash1 = nuc.find("-"); 
  int dash2;
  if (dash1 == npos)
    dash2 = npos;
  else
    dash2 = nuc.find("-", dash1+1);
  
  // nuc must be at least 4 characters or greater if it is in ZZLLAAAM form.
  if (nuc.length() >= 5 && dash1 != npos && dash2 != npos) {
    // Nuclide most likely in ZZLLAAAM Form, only form that contains two "-"'s.
    std::string zz = nuc.substr(0, dash1);
    std::string ll = nuc.substr(dash1+1, dash2);
    int zz_int = to_int(zz);
    // Verifying that the LL and ZZ point to the same element as secondary
    if(znum(ll) != zz_int)
      throw NotANuclide(nuc, "mismatched znum and chemical symbol");
    return zzllaaam_to_id(nuc);
  }

  // Get the string into a regular form
  std::string nucstr = pyne::to_upper(nuc);
  nucstr = pyne::remove_substring(nucstr, "-");
  int nuclen = nucstr.length();

  if (pyne::contains_substring(pyne::digits, nucstr.substr(0, 1))) {
    if (pyne::contains_substring(pyne::digits, nucstr.substr(nuclen-1, nuclen))) {
      // Nuclide must actually be an integer that 
      // just happens to be living in string form.
      newnuc = pyne::to_int(nucstr);
      newnuc = id(newnuc);
    } else {
      // probably in NIST-like form (242Am)
      // Here we know we have both digits and letters
      std::string anum_str = pyne::remove_characters(nucstr, pyne::alphabet);
      newnuc = pyne::to_int(anum_str) * 10000;

      // Add the Z-number
      elem_name = pyne::remove_characters(nucstr, pyne::digits);
      elem_name = pyne::capitalize(elem_name);
      if (0 < name_zz.count(elem_name))
        newnuc = (10000000 * name_zz[elem_name]) + newnuc;
      else
        throw NotANuclide(nucstr, newnuc);
    };
  } else if (pyne::contains_substring(pyne::alphabet, nucstr.substr(0, 1))) {
    // Nuclide is probably in name form, or some variation therein
    std::string anum_str = pyne::remove_characters(nucstr, pyne::alphabet);

    // natural element form, a la 'U' -> 920000000
    if (anum_str.empty()) {
      elem_name = pyne::capitalize(nucstr);    
      if (0 < name_zz.count(elem_name))
        return 10000000 * name_zz[elem_name]; 
    }

    int anum = pyne::to_int(anum_str);

    // bad form
    if (anum < 0)
      throw NotANuclide(nucstr, anum); 

    // Figure out if we are meta-stable or not
    std::string end_char = pyne::last_char(nucstr);
    if (end_char == "M")
      newnuc = (10000 * anum) + 1;
    else if (pyne::contains_substring(pyne::digits, end_char))
      newnuc = (10000 * anum);
    else
      throw NotANuclide(nucstr, newnuc);

    // Add the Z-number
    elem_name = pyne::remove_characters(nucstr.substr(0, nuclen-1), pyne::digits);
    elem_name = pyne::capitalize(elem_name);
    if (0 < name_zz.count(elem_name))
      newnuc = (10000000 * name_zz[elem_name]) + newnuc;
    else
      throw NotANuclide(nucstr, newnuc);
  } else {
    // Clearly not a nuclide
    throw NotANuclide(nuc, nucstr);
  }
  return newnuc;  
};


/***************************/
/*** iselement functions ***/
/***************************/

bool pyne::nucname::iselement(std::string nuc) {
  int n;
  try {
    n = id(nuc);
  }
  catch(NotANuclide) {
    return false;
  }
  return iselement(n);
};

bool pyne::nucname::iselement(const char * nuc) {
  return iselement(std::string(nuc));
};

bool pyne::nucname::iselement(int nuc) {
  int n;
  try {
    n = id(nuc);
  }
  catch(NotANuclide) {
    return false;
  }
 
  if (n <= 10000000)
    return false;
  int zzz = znum(n);
  int aaa = anum(n);
  if (zzz > 0 && aaa == 0)
    return true;  // is element
  return false;
};

/**********************/
/*** name functions ***/
/**********************/
std::string pyne::nucname::name(int nuc) {
  int nucid = id(nuc);
  std::string newnuc = "";

  int zzz = nucid / 10000000;
  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int aaa = aaassss / 10000;

  // Make sure the LL value is correct
  if (0 == zz_name.count(zzz))
    throw NotANuclide(nuc, nucid);

  // Add LL
  newnuc += zz_name[zzz];

  // Add A-number
  if (0 < aaa)
    newnuc += pyne::to_str(aaa);

  // Add meta-stable flag
  if (0 < ssss)
    newnuc += "M";

  return newnuc;
};



std::string pyne::nucname::name(const char * nuc) {
  std::string newnuc (nuc);
  return name(newnuc);
}


std::string pyne::nucname::name(std::string nuc) {
  return name(id(nuc));
}


/**********************/
/*** znum functions ***/
/**********************/
int pyne::nucname::znum(int nuc) {
  return id(nuc) / 10000000;
};

int pyne::nucname::znum(const char * nuc) {
  return id(nuc) / 10000000;
};

int pyne::nucname::znum(std::string nuc) {
  return id(nuc) / 10000000;
};

/**********************/
/*** anum functions ***/
/**********************/
int pyne::nucname::anum(int nuc) {
  return (id(nuc) / 10000) % 1000;
};

int pyne::nucname::anum(const char * nuc) {
  return (id(nuc) / 10000) % 1000;
};

int pyne::nucname::anum(std::string nuc) {
  return (id(nuc) / 10000) % 1000;
};

/**********************/
/*** snum functions ***/
/**********************/
int pyne::nucname::snum(int nuc) {
  return id(nuc) % 10000;
};

int pyne::nucname::snum(const char * nuc) {
  return id(nuc) % 10000;
};

int pyne::nucname::snum(std::string nuc) {
  return id(nuc) % 10000;
};

/************************/
/*** zzaaam functions ***/
/************************/
int pyne::nucname::zzaaam(int nuc) {
  int nucid = id(nuc);
  int zzzaaa = nucid / 10000;
  int ssss = nucid % 10000;
  if (10 <= ssss)
    ssss = 9;
  return zzzaaa*10 + ssss;
};


int pyne::nucname::zzaaam(const char * nuc) {
  std::string newnuc (nuc);
  return zzaaam(newnuc);
};


int pyne::nucname::zzaaam(std::string nuc) {
  return zzaaam(id(nuc));
};


int pyne::nucname::zzaaam_to_id(int nuc) {
  return (nuc/10)*10000 + (nuc%10);
};


int pyne::nucname::zzaaam_to_id(const char * nuc) {
  return zzaaam_to_id(std::string(nuc));
};


int pyne::nucname::zzaaam_to_id(std::string nuc) {
  return zzaaam_to_id(pyne::to_int(nuc));
};

/************************/
/*** zzzaaa functions ***/
/************************/
int pyne::nucname::zzzaaa(int nuc) {
  int nucid = id(nuc);
  int zzzaaa = nucid/10000;

  return zzzaaa;
};


int pyne::nucname::zzzaaa(const char * nuc) {
  std::string newnuc (nuc);
  return zzzaaa(newnuc);
};


int pyne::nucname::zzzaaa(std::string nuc) {
  return zzzaaa(id(nuc));
};


int pyne::nucname::zzzaaa_to_id(int nuc) {
  return (nuc)*10000;
};


int pyne::nucname::zzzaaa_to_id(const char * nuc) {
  return zzzaaa_to_id(std::string(nuc));
};


int pyne::nucname::zzzaaa_to_id(std::string nuc) {
  return zzzaaa_to_id(pyne::to_int(nuc));
};

/*************************/
/*** zzllaaam functions ***/
/*************************/
std::string pyne::nucname::zzllaaam(int nuc) {
  int nucid = id(nuc);
  std::string newnuc = "";

  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int zzz = nucid / 10000000;
  int aaa = aaassss / 10000;

  // Make sure the LL value is correct
  if (0 == zz_name.count(zzz))
    throw NotANuclide(nuc, nucid);
  //Adding ZZ
  newnuc += pyne::to_str(zzz);
  newnuc += "-";
  // Add LL
  newnuc += zz_name[zzz];
  // Add required dash
  newnuc += "-";
  // Add AAA
  if (0 < aaassss)
    newnuc += pyne::to_str(aaa);
  // Add meta-stable flag
  if (0 < ssss)
    newnuc += "m";
  return newnuc;
};


std::string pyne::nucname::zzllaaam(const char * nuc) {
  std::string newnuc (nuc);
  return zzllaaam(newnuc);
};


std::string pyne::nucname::zzllaaam(std::string nuc) {
  return zzllaaam(id(nuc));
};


int pyne::nucname::zzllaaam_to_id(const char * nuc) {
  return zzllaaam_to_id(std::string(nuc));
};


int pyne::nucname::zzllaaam_to_id(std::string nuc) {
  if (nuc.empty())
    throw NotANuclide(nuc, "<empty>");
  int nucid;
  std::string elem_name;

  // Get the string into a regular form
  std::string nucstr = pyne::to_upper(nuc);
  // Removing first two characters (redundant), for 1 digit nuclides, such
  // as 2-He-4, the first slash will be removed, and the second attempt to
  // remove the second slash will do nothing.  
  nucstr.erase(0,2);
  nucstr = pyne::remove_substring(nucstr, "-");
  // Does nothing if nuclide is short, otherwise removes the second "-" instance
  nucstr = pyne::remove_substring(nucstr, "-");
  int nuclen = nucstr.length();

  // Nuclide is probably in name form, or some variation therein
  std::string anum_str = pyne::remove_characters(nucstr, pyne::alphabet);

  // natural element form, a la 'U' -> 920000000
  if (anum_str.empty() || pyne::contains_substring(nucstr, "NAT")) {
    elem_name = pyne::capitalize(pyne::remove_substring(nucstr, "NAT")); 
    if (0 < name_zz.count(elem_name))
      return 10000000 * name_zz[elem_name]; 
  }
  int anum = pyne::to_int(anum_str);

  // Figure out if we are meta-stable or not
  std::string end_char = pyne::last_char(nucstr);
  if (end_char == "M")
    nucid = (10000 * anum) + 1;
  else if (pyne::contains_substring(pyne::digits, end_char))
    nucid = (10000 * anum);
  else
    throw NotANuclide(nucstr, nucid);

  // Add the Z-number
  elem_name = pyne::remove_characters(nucstr.substr(0, nuclen-1), pyne::digits);
  elem_name = pyne::capitalize(elem_name);
  if (0 < name_zz.count(elem_name))
    nucid = (10000000 * name_zz[elem_name]) + nucid;
  else
    throw NotANuclide(nucstr, nucid);
  return nucid;
};

/**********************/
/*** mcnp functions ***/
/**********************/
int pyne::nucname::mcnp(int nuc) {
  nuc = id(nuc);
  int ssss = nuc % 10000;
  int newnuc = nuc / 10000;

  // special case Am242(m)
  if (newnuc == 95242 && ssss < 2)
    ssss = (ssss + 1) % 2;

  // Handle the crazy MCNP meta-stable format
  if (0 != ssss && ssss < 10) 
    newnuc += 300 + (ssss * 100);

  return newnuc;
};



int pyne::nucname::mcnp(const char * nuc) {
  std::string newnuc (nuc);
  return mcnp(newnuc);
};



int pyne::nucname::mcnp(std::string nuc) {
  return mcnp(id(nuc));
};

//
// MCNP -> id
//
int pyne::nucname::mcnp_to_id(int nuc) {
  int zzz = nuc / 1000;
  int aaa = nuc % 1000; 
  if (zzz == 0)
    throw NotANuclide(nuc, "not in the MCNP format");
  else if (zzz <= aaa) {
    if (aaa - 400 < 0) {
      if (nuc == 95242)
        return nuc * 10000 + 1;  // special case MCNP Am-242m
      else
        return nuc * 10000;  // Nuclide in normal MCNP form
    } else {
      // Nuclide in MCNP metastable form
      if (nuc == 95642)
        return (95642 - 400)*10000;  // special case MCNP Am-242
      nuc = ((nuc - 400) * 10000) + 1;
      while (3.0 < (float ((nuc/10000)%1000) / float (nuc/10000000)))
        nuc -= 999999;
      return nuc;
    }
  } else if (aaa == 0)
    // MCNP form natural nuclide
    return zzz * 10000000;
  throw IndeterminateNuclideForm(nuc, "");
};


int pyne::nucname::mcnp_to_id(const char * nuc) {
  return mcnp_to_id(std::string(nuc));
};


int pyne::nucname::mcnp_to_id(std::string nuc) {
  return mcnp_to_id(pyne::to_int(nuc));
};


/**********************/
/*** fluka functions ***/
/**********************/
std::string pyne::nucname::fluka(int nuc) {
  int x = id(nuc);
  if (zz_fluka.count(x) == 0) {
    throw NotANuclide(nuc, "fluka name could not be found");
  }
  return zz_fluka[x];
};


//
// FLUKA name -> id
//
int pyne::nucname::fluka_to_id(std::string name) {
  if (fluka_zz.count(name) == 0) {
    throw NotANuclide(-1, "No nuclide: fluka name could not be found");
  }
  return fluka_zz[name];
}

int pyne::nucname::fluka_to_id(char * name) {
  return fluka_to_id(std::string(name));
}


/*************************/
/*** serpent functions ***/
/*************************/
std::string pyne::nucname::serpent(int nuc) {
  int nucid = id(nuc);
  std::string newnuc = "";

  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int zzz = nucid / 10000000;
  int aaa = aaassss / 10000;

  // Make sure the LL value is correct
  if (0 == zz_name.count(zzz))
    throw NotANuclide(nuc, nucid);

  // Add LL
  std::string llupper = pyne::to_upper(zz_name[zzz]);
  std::string lllower = pyne::to_lower(zz_name[zzz]);
  newnuc += llupper[0];
  for (int l = 1; l < lllower.size(); l++)
    newnuc += lllower[l];  

  // Add required dash
  newnuc += "-";

  // Add A-number
  if (0 < aaassss)
    newnuc += pyne::to_str(aaa);
  else if (0 == aaassss)
    newnuc += "nat";

  // Add meta-stable flag
  if (0 < ssss)
    newnuc += "m";

  return newnuc;
};


std::string pyne::nucname::serpent(const char * nuc) {
  std::string newnuc (nuc);
  return serpent(newnuc);
};


std::string pyne::nucname::serpent(std::string nuc) {
  return serpent(id(nuc));
};

//
// Serpent -> id
//
//int pyne::nucname::serpent_to_id(int nuc)
//{
// Should be ZAID
//};


int pyne::nucname::serpent_to_id(const char * nuc) {
  return serpent_to_id(std::string(nuc));
};


int pyne::nucname::serpent_to_id(std::string nuc) {
  if (nuc.empty())
    throw NotANuclide(nuc, "<empty>");
  int nucid;
  std::string elem_name;

  // Get the string into a regular form
  std::string nucstr = pyne::to_upper(nuc);
  nucstr = pyne::remove_substring(nucstr, "-");
  int nuclen = nucstr.length();

  // Nuclide is probably in name form, or some variation therein
  std::string anum_str = pyne::remove_characters(nucstr, pyne::alphabet);

  // natural element form, a la 'U' -> 920000000
  if (anum_str.empty() || pyne::contains_substring(nucstr, "NAT")) {
    elem_name = pyne::capitalize(pyne::remove_substring(nucstr, "NAT")); 
    if (0 < name_zz.count(elem_name))
      return 10000000 * name_zz[elem_name]; 
  }
  int anum = pyne::to_int(anum_str);

  // Figure out if we are meta-stable or not
  std::string end_char = pyne::last_char(nucstr);
  if (end_char == "M")
    nucid = (10000 * anum) + 1;
  else if (pyne::contains_substring(pyne::digits, end_char))
    nucid = (10000 * anum);
  else
    throw NotANuclide(nucstr, nucid);

  // Add the Z-number
  elem_name = pyne::remove_characters(nucstr.substr(0, nuclen-1), pyne::digits);
  elem_name = pyne::capitalize(elem_name);
  if (0 < name_zz.count(elem_name))
    nucid = (10000000 * name_zz[elem_name]) + nucid;
  else
    throw NotANuclide(nucstr, nucid);
  return nucid;
};


/**********************/
/*** nist functions ***/
/**********************/
std::string pyne::nucname::nist(int nuc) {
  int nucid = id(nuc);
  std::string newnuc = "";

  int zzz = nucid / 10000000;
  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int aaa = aaassss / 10000;

  // Make sure the LL value is correct
  if (0 == zz_name.count(zzz))
    throw NotANuclide(nuc, nucid);

  // Add A-number
  if (0 < aaassss)
    newnuc += pyne::to_str(aaa);

  // Add name
  std::string name_upper = pyne::to_upper(zz_name[zzz]);
  std::string name_lower = pyne::to_lower(zz_name[zzz]);
  newnuc += name_upper[0];
  for (int l = 1; l < name_lower.size(); l++)
    newnuc += name_lower[l];  

  // Add meta-stable flag
  // No metastable flag for NIST, 
  // but could add star, by uncommenting below
  //if (0 < mod_10)
  //  newnuc += "*";

  return newnuc;
};


std::string pyne::nucname::nist(const char * nuc) {
  std::string newnuc (nuc);
  return nist(newnuc);
};


std::string pyne::nucname::nist(std::string nuc) {
  return nist(id(nuc));
};


//
// NIST -> id
//
//int pyne::nucname::nist_to_id(int nuc)
//{
// NON-EXISTANT
//};

int pyne::nucname::nist_to_id(const char * nuc) {
  return nist_to_id(std::string(nuc));
};

int pyne::nucname::nist_to_id(std::string nuc) {
  if (nuc.empty())
    throw NotANuclide(nuc, "<empty>");
  int nucid;
  nuc = pyne::to_upper(nuc);
  std::string elem_name;
  int nuclen = nuc.length();

  // Nuclide is probably in name form, or some variation therein
  std::string anum_str = pyne::remove_characters(nuc, pyne::alphabet);

  // natural element form, a la 'U' -> 920000000
  if (anum_str.empty()) {
    elem_name = pyne::capitalize(nuc);
    if (0 < name_zz.count(elem_name))
      return 10000000 * name_zz[elem_name]; 
  }
  nucid = pyne::to_int(anum_str) * 10000;

  // Add the Z-number
  elem_name = pyne::remove_characters(nuc, pyne::digits);
  elem_name = pyne::capitalize(elem_name);
  if (0 < name_zz.count(elem_name))
    nucid = (10000000 * name_zz[elem_name]) + nucid;
  else
    throw NotANuclide(nuc, nucid);
  return nucid;
};


/************************/
/*** cinder functions ***/
/************************/
int pyne::nucname::cinder(int nuc) {
  // cinder nuclides of form aaazzzm
  int nucid = id(nuc);
  int zzz = nucid / 10000000;
  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int aaa = aaassss / 10000;
  if (10 <= ssss)
    ssss = 9;
  return (aaa*10000) + (zzz*10) + ssss;
};



int pyne::nucname::cinder(const char * nuc) {
  std::string newnuc (nuc);
  return cinder(newnuc);
};



int pyne::nucname::cinder(std::string nuc) {
  return cinder(id(nuc));
};

//
// Cinder -> Id
//
int pyne::nucname::cinder_to_id(int nuc) {
  int ssss = nuc % 10;
  int aaazzz = nuc / 10;
  int zzz = aaazzz % 1000;
  int aaa = aaazzz / 1000;
  return (zzz * 10000000) + (aaa * 10000) + ssss;
};


int pyne::nucname::cinder_to_id(const char * nuc) {
  return cinder_to_id(std::string(nuc));
};


int pyne::nucname::cinder_to_id(std::string nuc) {
  return cinder_to_id(pyne::to_int(nuc));
};




/**********************/
/*** ALARA functions ***/
/**********************/
std::string pyne::nucname::alara(int nuc) {
  int nucid = id(nuc);
  std::string newnuc = "";
  std::string ll = "";

  int zzz = nucid / 10000000;
  int ssss = nucid % 10000;
  int aaassss = nucid % 10000000;
  int aaa = aaassss / 10000;

  // Make sure the LL value is correct
  if (0 == zz_name.count(zzz))
    throw NotANuclide(nuc, nucid);

  // Add LL, in lower case
  ll += zz_name[zzz];

  for(int i = 0; ll[i] != '\0'; i++)
    ll[i] = tolower(ll[i]);
  newnuc += ll;

  // Add A-number
  if (0 < aaassss){
    newnuc += ":";
    newnuc += pyne::to_str(aaa);
  }

  // Note, ALARA input format does not use metastable flag
  return newnuc;
};


std::string pyne::nucname::alara(const char * nuc) {
  std::string newnuc (nuc);
  return alara(newnuc);
}


std::string pyne::nucname::alara(std::string nuc) {
  return alara(id(nuc));
}


//
// Cinder -> Id
//
//int pyne::nucname::alara_to_id(int nuc)
//{
// Not Possible
//};


int pyne::nucname::alara_to_id(const char * nuc) {
  return alara_to_id(std::string(nuc));
};


int pyne::nucname::alara_to_id(std::string nuc) {
  if (nuc.empty())
    throw NotANuclide(nuc, "<empty>");
  int nucid;
  nuc = pyne::to_upper(pyne::remove_characters(nuc, ":"));
  std::string elem_name;
  int nuclen = nuc.length();

  // Nuclide is probably in name form, or some variation therein
  std::string anum_str = pyne::remove_characters(nuc, pyne::alphabet);

  // natural element form, a la 'U' -> 920000000
  if (anum_str.empty()) {
    elem_name = pyne::capitalize(nuc);
    if (0 < name_zz.count(elem_name))
      return 10000000 * name_zz[elem_name]; 
  }
  nucid = pyne::to_int(anum_str) * 10000;

  // Add the Z-number
  elem_name = pyne::remove_characters(nuc, pyne::digits);
  elem_name = pyne::capitalize(elem_name);
  if (0 < name_zz.count(elem_name))
    nucid = (10000000 * name_zz[elem_name]) + nucid;
  else
    throw NotANuclide(nuc, nucid);
  return nucid;
};




/***********************/
/***  SZA functions  ***/
/***********************/
int pyne::nucname::sza(int nuc) {
  int nucid = id(nuc);
  int zzzaaa = nucid / 10000;
  int sss = nucid % 10000;
  return sss * 1000000 + zzzaaa;
}


int pyne::nucname::sza(const char * nuc) {
  std::string newnuc (nuc);
  return sza(newnuc);
}


int pyne::nucname::sza(std::string nuc) {
  return sza(id(nuc));
}


int pyne::nucname::sza_to_id(int nuc) {
  int sss = nuc / 1000000;
  int zzzaaa = nuc % 1000000;
  if (5 < sss){
  // Unphysical metastable state warning 
   warning("You have indicated a metastable state of " + pyne::to_str(sss) + ". Metastable state above 5, possibly unphysical. ");
  }
  return zzzaaa * 10000 + sss;
}


int pyne::nucname::sza_to_id(const char * nuc) {
  std::string newnuc (nuc);
  return sza_to_id(newnuc);
}


int pyne::nucname::sza_to_id(std::string nuc) {
  return sza_to_id(pyne::to_int(nuc));
}


void pyne::nucname::_load_state_map(){
    for (int i = 0; i < TOTAL_STATE_MAPS; ++i) {
       state_id_map[map_nuc_ids[i]] = map_metastable[i];
    }
}

int pyne::nucname::state_id_to_id(int state) {
    int zzzaaa = (state / 10000) * 10000;
    int state_number = state % 10000;
    if (state_number == 0) return state;
    std::map<int, int>::iterator nuc_iter, nuc_end;

    nuc_iter = state_id_map.find(state);
    nuc_end = state_id_map.end();
    if (nuc_iter != nuc_end){ 
     int m = (*nuc_iter).second;
     return zzzaaa + m;
    }        

    if (state_id_map.empty())  {
      _load_state_map();
      return state_id_to_id(state);
    }
    throw IndeterminateNuclideForm(state, "no matching metastable state");
}


int pyne::nucname::id_to_state_id(int nuc_id) {
    int zzzaaa = (nuc_id / 10000) * 10000;
    int state = nuc_id % 10000;
    if (state == 0) return nuc_id;
    std::map<int, int>::iterator nuc_iter, nuc_end, it;
    
    nuc_iter = state_id_map.lower_bound(nuc_id);
    nuc_end = state_id_map.upper_bound(nuc_id + 10000);
    for (it = nuc_iter; it!= nuc_end; ++it){
        if (state == it->second) {
          return it->first;
        }
    }
    int m = (*nuc_iter).second;
    
    if (state_id_map.empty())  {
      _load_state_map();
      return id_to_state_id(nuc_id);
    }
    throw IndeterminateNuclideForm(state, "no matching state id");
}


/************************/
/*** ENSDF functions ***/
/************************/
//
// ENSDF  -> Id
//

int pyne::nucname::ensdf_to_id(const char * nuc) {
  return ensdf_to_id(std::string(nuc));
};

int pyne::nucname::ensdf_to_id(std::string nuc) {
  if (nuc.size() < 4) {
    return nucname::id(nuc);
  } else if (std::isdigit(nuc[3])) {
    int aaa = to_int(nuc.substr(0, 3));
    int zzz;
    std::string xx_str = nuc.substr(3,2); 
    zzz = to_int(xx_str) + 100;
    int nid = 10000 * aaa + 10000000 * zzz;
    return nid;
  } else {
    return nucname::id(nuc);
  }
  
};

//
// end of src/nucname.cpp
//
