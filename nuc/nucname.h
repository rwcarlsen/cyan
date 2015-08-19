// This contents of this file were extracted from the pyne project - license
// below:
/*
Copyright 2011-2015, the PyNE Development Team. All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are
permitted provided that the following conditions are met:

   1. Redistributions of source code must retain the above copyright notice, this list of
      conditions and the following disclaimer.

   2. Redistributions in binary form must reproduce the above copyright notice, this list
      of conditions and the following disclaimer in the documentation and/or other materials
      provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE PYNE DEVELOPMENT TEAM ``AS IS'' AND ANY EXPRESS OR IMPLIED
WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> OR
CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF
ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation are those of the
authors and should not be interpreted as representing official policies, either expressed
or implied, of the stakeholders of the PyNE project or the employers of PyNE developers.

-------------------------------------------------------------------------------
The files cpp/measure.cpp and cpp/measure.hpp are covered by:

Copyright 2004 Sandia Corporation.  Under the terms of Contract
DE-AC04-94AL85000 with Sandia Coroporation, the U.S. Government
retains certain rights in this software.

http://trac.mcs.anl.gov/projects/ITAPS/wiki/MOAB
*/

#include <iostream>
#include <string>
#include <sstream>
#include <map>
#include <set>
#include <exception>
#include <stdlib.h>
#include <stdio.h>

namespace pyne
{

  // String Transformations
  /// string of digit characters
  static std::string digits = "0123456789";
  /// uppercase alphabetical characters
  static std::string alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
  /// string of all valid word characters for variable names in programing languages.
  static std::string words = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_";

  /// \name String Conversion Functions
  /// \{
  /// Converts the variables of various types to their C++ string representation.
  std::string to_str(int t);
  std::string to_str(unsigned int t);
  std::string to_str(double t);
  std::string to_str(bool t);
  /// \}

  int to_int(std::string s);  ///< Converts a string of digits to an int using atoi().

  double to_dbl(std::string s);  ///< Converts a valid string to a float using atof().

  /// Converts a string from ENDF format to a float. Only handles E-less format
  /// but is roughly 5 times faster than endftod.
  double endftod_cpp(char * s);
  double endftod_f(char * s); ///< Converts a string from ENDF format to a float.
  extern  double (*endftod)(char * s); ///< endftod function pointer. defaults to fortran

  void use_fast_endftod();/// switches endftod to fast cpp version

  /// Returns an all upper case copy of the string.
  std::string to_upper(std::string s);

  /// Returns an all lower case copy of the string.
  std::string to_lower(std::string s);

  /// Returns a capitalized copy of the string.
  std::string capitalize(std::string s);

  /// Finds and returns the first white-space delimited token of a line.
  /// \param line a character array to take the first token from.
  /// \param max_l an upper bound to the length of the token.  Must be 11 or less.
  /// \returns a the flag as a string
  std::string get_flag(char line[], int max_l);

  /// Creates a copy of \a s with all instances of \a substr taken out.
  std::string remove_substring(std::string s, std::string substr);

  /// Removes all characters in the string \a chars from \a s.
  std::string remove_characters(std::string s, std::string chars);

  /// Replaces all instance of \a substr in \a s with \a repstr.
  std::string replace_all_substrings(std::string s, std::string substr,
                                                    std::string repstr);

  /// Returns the last character in a string.
  std::string last_char(std::string s);

  /// Returns the slice of a string \a s using the negative index \a n and the
  /// length of the slice \a l.
  std::string slice_from_end(std::string s, int n=-1, int l=1);

  /// Returns true if \a a <= \a b <= \a c and flase otherwise.
  bool ternary_ge(int a, int b, int c);

  /// Returns true if \a substr is in \a s.
  bool contains_substring(std::string s, std::string substr);

  /// Calculates a version of the string \a name that is also a valid variable name.
  /// That is to say that the return value uses only word characters.
  std::string natural_naming(std::string name);

  /// Finds the slope of a line from the points (\a x1, \a y1) and (\a x2, \a y2).
  double slope (double x2, double y2, double x1, double y1);

  /// Solves the equation for the line y = mx + b, given \a x and the points that
  /// form the line: (\a x1, \a y1) and (\a x2, \a y2).
  double solve_line (double x, double x2, double y2, double x1, double y1);

  double tanh(double x);  ///< The hyperbolic tangent function.
  double coth(double x);  ///< The hyperbolic cotangent function.


  // Message Helpers
  extern bool USE_WARNINGS;
  /// Toggles warnings on and off
  bool toggle_warnings();

  /// Prints a warning message.
  void warning(std::string s);


//! Nuclide naming conventions
namespace nucname
{
  typedef std::string name_t; ///< name type
  typedef int zz_t;           ///< Z number type

  typedef std::map<name_t, zz_t> name_zz_t; ///< name and Z num map type
  typedef name_zz_t::iterator name_zz_iter; ///< name and Z num iter type
  name_zz_t get_name_zz();  ///< Creates standard name to Z number mapping.
  extern name_zz_t name_zz; ///< name to Z num map

  typedef std::map<zz_t, name_t> zzname_t;  ///< Z num to name map type
  typedef zzname_t::iterator zzname_iter;   ///< Z num to name iter type
  zzname_t get_zz_name();   ///< Creates standard Z number to name mapping.
  extern zzname_t zz_name;  ///< Z num to name map

  name_zz_t get_fluka_zz();  ///< Creates standard fluka-name to nucid mapping.
  extern name_zz_t fluka_zz; ///< fluka-name to nucid map
  zzname_t get_zz_fluka();   ///< Creates standard nucid to fluka-name mapping.
  extern zzname_t zz_fluka;  ///< nucid to fluka-name map
  /******************************************/
  /*** Define useful elemental group sets ***/
  /******************************************/

  /// name grouping type (for testing containment)
  typedef std::set<name_t> name_group;
  typedef name_group::iterator name_group_iter; ///< name grouping iter type

  /// Z number grouping type (for testing containment)
  typedef std::set<zz_t> zz_group;
  typedef zz_group::iterator zz_group_iter; ///< Z number grouping iter

  /// Converts a name group to a Z number group.
  /// \param eg a grouping of nuclides by name
  /// \return a Z numbered group
  zz_group name_to_zz_group (name_group eg);

  extern name_t LAN_array[15];  ///< array of lanthanide names
  extern name_group LAN;        ///< lanthanide name group
  extern zz_group lan;          ///< lanthanide Z number group

  extern name_t ACT_array[15];  ///< array of actinide names
  extern name_group ACT;        ///< actinide name group
  extern zz_group act;          ///< actinide Z number group

  extern name_t TRU_array[22];  ///< array of transuranic names
  extern name_group TRU;        ///< transuranic name group
  extern zz_group tru;          ///< transuranic Z number group

  extern name_t MA_array[10];   ///< array of minor actinide names
  extern name_group MA;         ///< minor actinide name group
  extern zz_group ma;           ///< minor actinide Z number group

  extern name_t FP_array[88];   ///< array of fission product names
  extern name_group FP;         ///< fission product name group
  extern zz_group fp;           ///< fission product Z number group


  /******************/
  /*** Exceptions ***/
  /******************/

  /// Custom expection for declaring that a value does not follow a recognizable
  /// nuclide naming convention.
  class NotANuclide : public std::exception
  {
  public:
    /// default constructor
    NotANuclide () {};

    /// default destructor
    ~NotANuclide () throw () {};

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    NotANuclide(std::string wasptr, std::string nowptr)
    {
       nucwas = wasptr;
       nucnow = nowptr;
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    NotANuclide(std::string wasptr, int nowptr)
    {
      nucwas = wasptr;
      nucnow = pyne::to_str(nowptr);
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    NotANuclide(int wasptr, std::string nowptr)
    {
      nucwas = pyne::to_str(wasptr);
      nucnow = nowptr;
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    NotANuclide(int wasptr, int nowptr)
    {
      nucwas = pyne::to_str(wasptr);
      nucnow = pyne::to_str(nowptr);
    };

    /// Generates an informational message for the exception
    /// \return The error string
    virtual const char* what() const throw()
    {
      std::string NaNEstr ("Not a Nuclide! ");
      if (!nucwas.empty())
        NaNEstr += nucwas;

      if (!nucnow.empty())
      {
        NaNEstr += " --> ";
        NaNEstr += nucnow;
      }
      return (const char *) NaNEstr.c_str();
    };

  private:
    std::string nucwas; ///< previous nuclide state
    std::string nucnow; ///< current nuclide state
  };

  /// Custom expection for declaring that a value represents one or more nuclides
  /// in one or more namig conventions
  class IndeterminateNuclideForm : public std::exception
  {
  public:
    /// default constructor
    IndeterminateNuclideForm () {};

    /// default destuctor
    ~IndeterminateNuclideForm () throw () {};

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    IndeterminateNuclideForm(std::string wasptr, std::string nowptr)
    {
       nucwas = wasptr;
       nucnow = nowptr;
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    IndeterminateNuclideForm(std::string wasptr, int nowptr)
    {
      nucwas = wasptr;
      nucnow = pyne::to_str(nowptr);
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    IndeterminateNuclideForm(int wasptr, std::string nowptr)
    {
      nucwas = pyne::to_str(wasptr);
      nucnow = nowptr;
    };

    /// Constructor given previous and current state of nulide name
    /// \param wasptr Previous state, typically user input.
    /// \param nowptr Current state, as far as PyNE could get.
    IndeterminateNuclideForm(int wasptr, int nowptr)
    {
      nucwas = pyne::to_str(wasptr);
      nucnow = pyne::to_str(nowptr);
    };

    /// Generates an informational message for the exception
    /// \return The error string
    virtual const char* what() const throw()
    {
      std::string INFEstr ("Indeterminate nuclide form: ");
      if (!nucwas.empty())
        INFEstr += nucwas;

      if (!nucnow.empty())
      {
        INFEstr += " --> ";
        INFEstr += nucnow;
      }
      return (const char *) INFEstr.c_str();
    }

  private:
    std::string nucwas; ///< previous nuclide state
    std::string nucnow; ///< current nuclide state
  };

  /// \name isnuclide functions
  /// \{
  /// These functions test if an input \a nuc is a valid nuclide.
  /// \param nuc a possible nuclide
  /// \return a bool
  bool isnuclide(std::string nuc);
  bool isnuclide(const char * nuc);
  bool isnuclide(int nuc);
  /// \}

  /// \name iselement functions
  /// \{
  /// These functions test if an input \a nuc is a valid element.
  /// \param nuc a possible element
  /// \return a bool
  bool iselement(std::string nuc);
  bool iselement(const char * nuc);
  bool iselement(int nuc);

  /// \}
  /// \name Identifier Form Functions
  /// \{
  /// The 'id' nuclide naming convention is the canonical form for representing
  /// nuclides in PyNE. This is termed a ZAS, or ZZZAAASSSS, representation because
  /// It stores 3 Z-number digits, 3 A-number digits, followed by 4 S-number digits
  /// which the nucleus excitation state.
  ///
  /// The id() function will always return an nuclide in id form, if successful.
  /// If the input nuclide is in id form already, then this is function does no
  /// work. For all other formats, the id() function provides a best-guess based
  /// on a heirarchy of other formats that is used to resolve ambiguities between
  /// naming conventions. For integer input the form resolution order is:
  ///   - id
  ///   - zz (elemental z-num only given)
  ///   - zzaaam
  ///   - cinder (aaazzzm)
  ///   - mcnp
  ///   - zzaaa
  /// For string (or char *) input the form resolution order is as follows:
  ///   - ZZ-LL-AAAM
  ///   - Integer form in a string representation, uses interger resolution
  ///   - NIST
  ///   - name form
  ///   - Serpent
  ///   - LL (element symbol)
  /// For well-defined situations where you know ahead of time what format the
  /// nuclide is in, you should use the various form_to_id() functions, rather
  /// than the id() function which is meant to resolve possibly ambiquous cases.
  /// \param nuc a nuclide
  /// \return nucid 32-bit integer identifier
  int id(int nuc);
  int id(const char * nuc);
  int id(std::string nuc);
  /// \}

  /// \name Name Form Functions
  /// \{
  /// The 'name' nuclide naming convention is the more common, human readable
  /// notation. The chemical symbol (one or two characters long) is first, followed
  /// by the nucleon number. Lastly if the nuclide is metastable, the letter M is
  /// concatenated to the end. For example, ‘H-1’ and ‘Am242M’ are both valid.
  /// Note that nucname will always return name form with dashes removed, the
  /// chemical symbol used for letter casing (ie 'Pu'), and a trailing upercase 'M'
  /// for a metastable flag. The name() function first converts functions to id form
  /// using the id() function. Thus the form order resolution for id() also applies
  /// here.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  std::string name(int nuc);
  std::string name(const char * nuc);
  std::string name(std::string nuc);
  /// \}

  /// \name Z-Number Functions
  /// \{
  /// The Z-number, or charge number, represents the number of protons in a
  /// nuclide.  This function returns that number.
  /// \param nuc a nuclide
  /// \return an integer Z-number.
  int znum(int nuc);
  int znum(const char * nuc);
  int znum(std::string nuc);
  /// \}

  /// \name A-Number Functions
  /// \{
  /// The A-number, or nucleon number, represents the number of protons and
  /// neutrons in a nuclide.  This function returns that number.
  /// \param nuc a nuclide
  /// \return an integer A-number.
  int anum(int nuc);
  int anum(const char * nuc);
  int anum(std::string nuc);
  /// \}

  /// \name S-Number Functions
  /// \{
  /// The S-number, or excitation state number, represents the excitation
  /// level of a nuclide.  Normally, this is zero.  This function returns
  /// that number.
  /// \param nuc a nuclide
  /// \return an integer A-number.
  int snum(int nuc);
  int snum(const char * nuc);
  int snum(std::string nuc);
  /// \}

  /// \name ZZAAAM Form Functions
  /// \{
  /// The ZZAAAM nuclide naming convention is the former canonical form for
  /// nuclides in PyNE. This places the charge of the nucleus out front, then has
  /// three digits for the atomic mass number, and ends with a metastable flag
  /// (0 = ground, 1 = first excited state, 2 = second excited state, etc).
  /// Uranium-235 here would be expressed as ‘922350’.
  /// \param nuc a nuclide
  /// \return an integer nuclide identifier.
  int zzaaam(int nuc);
  int zzaaam(const char * nuc);
  int zzaaam(std::string nuc);
  /// \}

  /// \name ZZAAAM Form to Identifier Form Functions
  /// \{
  /// This converts from the ZZAAAM nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in ZZAAAM form.
  /// \return an integer id nuclide identifier.
  int zzaaam_to_id(int nuc);
  int zzaaam_to_id(const char * nuc);
  int zzaaam_to_id(std::string nuc);
  /// \}


  /// \name ZZZAAA Form Functions
  /// \{
  /// The ZZZAAA nuclide naming convention is a form in which the nuclides three
  ///digit ZZZ number is followed by the 3 digit AAA number.  If the ZZZ number
  ///is 2 digits, the preceding zeros are not included.
  /// Uranium-235 here would be expressed as ‘92235’.
  /// \param nuc a nuclide
  /// \return an integer nuclide identifier.
  int zzzaaa(int nuc);
  int zzzaaa(const char * nuc);
  int zzzaaa(std::string nuc);
  /// \}


  /// \name ZZZAAA Form to Identifier Form Functions
  /// \{
  /// This converts from the ZZZAAA nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in ZZZAAA form.
  /// \return an integer id nuclide identifier.
  int zzzaaa_to_id(int nuc);
  int zzzaaa_to_id(const char * nuc);
  int zzzaaa_to_id(std::string nuc);
  /// \}


  /// \name ZZLLAAAM Form Functions
  /// \{
  /// The ZZLLAAAM nuclide naming convention is a form in which the nuclides
  /// AA number is followed by the redundant two LL characters, followed by
  /// the nuclides ZZZ number.  Can also be followed with a metastable flag.
  /// Uranium-235 here would be expressed as ‘92-U-235’.
  /// \param nuc a nuclide
  /// \return an integer nuclide identifier.
  std::string zzllaaam(int nuc);
  std::string zzllaaam(const char * nuc);
  std::string zzllaaam(std::string nuc);
  /// \}


  /// \name ZZLLAAAM Form to Identifier Form Functions
  /// \{
  /// This converts from the ZZLLAAAM nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in ZZLLAAAM form.
  /// \return an integer id nuclide identifier.
  //int zzllaaam_to_id(int nuc);
  int zzllaaam_to_id(const char * nuc);
  int zzllaaam_to_id(std::string nuc);
  /// \}


  /// \name MCNP Form Functions
  /// \{
  /// This is the naming convention used by the MCNP suite of codes.
  /// The MCNP format for entering nuclides is unfortunately non-standard.
  /// In most ways it is similar to zzaaam form, except that it lacks the metastable
  /// flag. For information on how metastable isotopes are named, please consult the
  /// MCNP documentation for more information.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  int mcnp(int nuc);
  int mcnp(const char * nuc);
  int mcnp(std::string nuc);
  /// \}

  /// \name MCNP Form to Identifier Form Functions
  /// \{
  /// This converts from the MCNP nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in MCNP form.
  /// \return an integer id nuclide identifier.
  int mcnp_to_id(int nuc);
  int mcnp_to_id(const char * nuc);
  int mcnp_to_id(std::string nuc);
  /// \}

  /// \name FLUKA Form Functions
  /// \{
  /// This is the naming convention used by the FLUKA suite of codes.
  /// The FLUKA format for entering nuclides requires some knowledge of FLUKA
  /// The nuclide in must cases should be the atomic # times 10000000.
  /// The exceptions are for FLUKA's named isotopes
  /// See the FLUKA Manual for more information.
  /// \param nuc a nuclide
  /// \return the received FLUKA name
  std::string fluka(int nuc);
  /// \}

  /// \name FLUKA Form to Identifier Form Functions
  /// \{
  /// This converts from the FLUKA name to the
  /// id canonical form  for nuclides in PyNE.
  /// \param name a fluka name
  /// \return an integer id nuclide identifier.
  int fluka_to_id(std::string name);
  int fluka_to_id(char * name);
  /// \}

  /// \name Serpent Form Functions
  /// \{
  /// This is the string-based naming convention used by the Serpent suite of codes.
  /// The serpent naming convention is similar to name form. However, only the first
  /// letter in the chemical symbol is uppercase, the dash is always present, and the
  /// the meta-stable flag is lowercase. For instance, ‘Am-242m’ is the valid serpent
  /// notation for this nuclide.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  std::string serpent(int nuc);
  std::string serpent(const char * nuc);
  std::string serpent(std::string nuc);
  /// \}

  /// \name Serpent Form to Identifier Form Functions
  /// \{
  /// This converts from the Serpent nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in Serpent form.
  /// \return an integer id nuclide identifier.
  //int serpent_to_id(int nuc);  Should be ZAID
  int serpent_to_id(const char * nuc);
  int serpent_to_id(std::string nuc);
  /// \}

  /// \name NIST Form Functions
  /// \{
  /// This is the string-based naming convention used by NIST.
  /// The NIST naming convention is also similar to the Serpent form. However, this
  /// convention contains no metastable information. Moreover, the A-number comes
  /// before the element symbol. For example, ‘242Am’ is the valid NIST notation.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  std::string nist(int nuc);
  std::string nist(const char * nuc);
  std::string nist(std::string nuc);
  /// \}

  /// \name NIST Form to Identifier Form Functions
  /// \{
  /// This converts from the NIST nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in NIST form.
  /// \return an integer id nuclide identifier.
  //int serpent_to_id(int nuc);  NON-EXISTANT
  int nist_to_id(const char * nuc);
  int nist_to_id(std::string nuc);
  /// \}

  /// \name CINDER Form Functions
  /// \{
  /// This is the naming convention used by the CINDER burnup library.
  /// The CINDER format is similar to zzaaam form except that the placement of the
  /// Z- and A-numbers are swapped. Therefore, this format is effectively aaazzzm.
  /// For example, ‘2420951’ is the valid cinder notation for ‘AM242M’.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  int cinder(int nuc);
  int cinder(const char * nuc);
  int cinder(std::string nuc);
  /// \}

  /// \name Cinder Form to Identifier Form Functions
  /// \{
  /// This converts from the Cinder nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in Cinder form.
  /// \return an integer id nuclide identifier.
  int cinder_to_id(int nuc);
  int cinder_to_id(const char * nuc);
  int cinder_to_id(std::string nuc);
  /// \}

  /// \name ALARA Form Functions
  /// \{
  /// This is the format used in the ALARA activation code elements library.
  /// For elements, the form is "ll" where ll is the atomic symbol. For isotopes
  /// the form is "ll:AAA". No metastable isotope flag is used.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  std::string alara(int nuc);
  std::string alara(const char * nuc);
  std::string alara(std::string nuc);
  /// \}

  /// \name ALARA Form to Identifier Form Functions
  /// \{
  /// This converts from the ALARA nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in ALARA form.
  /// \return an integer id nuclide identifier.
  //int alara_to_id(int nuc); NOT POSSIBLE
  int alara_to_id(const char * nuc);
  int alara_to_id(std::string nuc);
  /// \}

  /// \name SZA Form Functions
  /// \{
  /// This is the new format for ACE data tables in the form SSSZZZAAA.
  /// The first three digits represent the excited state (000 = ground,
  /// 001 = first excited state, 002 = second excited state, etc).
  /// The second three digits are the atomic number and the last three
  /// digits are the atomic mass. Prepending zeros can be omitted, making
  /// the SZA form equal to the MCNP form for non-excited nuclides.
  /// \param nuc a nuclide
  /// \return a string nuclide identifier.
  int sza(int nuc);
  int sza(const char * nuc);
  int sza(std::string nuc);
  /// \}

  /// \name SZA Form to Identifier Form Functions
  /// \{
  /// This converts from the SZA nuclide naming convention
  /// to the id canonical form  for nuclides in PyNE.
  /// \param nuc a nuclide in SZA form.
  /// \return an integer id nuclide identifier.
  int sza_to_id(int nuc);
  int sza_to_id(const char * nuc);
  int sza_to_id(std::string nuc);
  /// \}

  /// \name Ground State Form Functions
  /// \{
  /// This form stores the nuclide in id form, but removes
  /// the state information about the nuclide.  I is in the same
  /// form as ID, but the four last digits are all zeros.
  /// \param nuc a nuclide
  /// \return a integer groundstate id
  inline int groundstate(int nuc) {return (id(nuc) / 10000 ) * 10000;};
  inline int groundstate(std::string nuc) {return groundstate(id(nuc));};
  inline int groundstate(const char * nuc) {return groundstate(std::string(nuc));};
  /// \}

  /// \name State Map functions
  /// \{
  /// These convert from/to decay state ids (used in decay data)
  /// to metastable ids (the PyNE default)
  void _load_state_map();
  int state_id_to_id(int state);
  int id_to_state_id(int nuc_id);
  extern std::map<int, int> state_id_map;
  /// \}

  /// \name ENSDF Form Functions
  /// \{
  /// This converts id's stored using standard ensdf syntax to nuc_id's
  /// \param ensdf nuc string
  /// \return PyNE nuc_id
  int ensdf_to_id(const char * nuc);
  int ensdf_to_id(std::string nuc);
  /// \}

};
};
