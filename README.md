# USAGE

```
  mzrecal [options] <mzMLfile>

  This program can be used to recalibrate MS data in an mzML file
  using peptide identifications in an accompanying mzID file.

  Recalibration is divided in 2 steps:
  1) Computation of recalibration coefficients. The coefficients are stored
     in a JSON file.
     This step reads an mzML file and mzID file, matches measured peaks to
     computed m/z values and computes recalibration coefficents using a method
     that is usefull for the instrument type. The instrument type (and other
     relavant values) are determined from the CV terms in the input files.
  2) Creating a recalibrated version of the MS file.
     This step reads the mzML file and JSON file with recalibration values,
     computes recalibrated m/z values for all peaks in spectra for which
     a valid recalibration was found, and writes a recalibrated mzML file.

  The default operation is computation of the recalibration values (the
  first step). Flag -recal switches to creation of the recalibrated mzML
  file (the second step).

OPTIONS:
  -cal string
        filename for output of computed calibration parameters
    
  -charge string
        charge range of calibrants, or the string "ident". If set to "ident",
        only the charge as found in the mzIdentMl file will be used for calibration.
         (default "1:5")
  -debug string
        Print debug output for given spectrum range e.g. 3:6
  -empty-non-calibrated
        Empty MS2 spectra for which the precursor was not recalibrated.
  -func string
        recalibration function to apply. If empty, a suitable
        function is determined from the instrument specified in the mzML file.
        Valid function names:
            FTICR, TOF, Orbitrap: Calibration function suitable for these instruments.
            POLY<N>: Polynomial with degee <N> (range 1:5)
  -minPeak float
        minimum intensity of peaks to be considered for recalibrating
         (default 10000)
  -mincals int
        minimum number of calibrants a spectrum should have to be recalibrated.
        If 0, the minimum number of calibrants is set to the smallest number needed
        for the choosen recalibration function plus one. In any other case, is the
        specified number is too low for the calibration function, it is increased to
        the minimum needed value.
  -mzAccept float
        max mz error (ppm) for accepting a calibrant for calibration
         (default 2)
  -mzTry float
        max mz error (ppm) for trying to use calibrant for calibration
         (default 10)
  -mzid string
        mzIdentMl filename
    
  -mzmlOut string
        recalibrated mzML filename (only together with -recal)
    
  -recal
        Switch between computation of recalibration parameters (default) and actual
                recalibration
  -rt string
        rt window (s)
         (default "-10.0:10.0")
  -scoreFilter string
        filter for PSM scores to accept. Format:
        <CVterm1|scorename1>([<minscore1>]:[<maxscore1>])...
        When multiple score names/CV terms are specified, the first one on the list
        that matches a score in the input file will be used.
        TODO: The default contains reasonable values for some common search engines
        and post-search scoring software:
        MS:1002257 (Comet:expectation value)
        MS:1001159 (SEQUEST:expectation value)
        MS:1002466 (PeptideShaker PSM score)
         (default "MS:1002466(0.99:)MS:1002257(0.0:1e-2)MS:1001159(0.0:1e-2)")

BUILD-IN CALIBRANTS:
  In addition to the identified peptides, mzrecal will also use
  for recalibration a number of compounds that are commonly found in many
  samples. These compound are all assumed to have +1 charge. The following
  list shows the build-in compounds with their (uncharged) masses:
     cyclosiloxane6 (444.112748)
     cyclosiloxane7 (518.131539)
     cyclosiloxane8 (592.150331)
     cyclosiloxane9 (666.169122)
     cyclosiloxane10 (740.187913)
     cyclosiloxane11 (814.206705)
     cyclosiloxane12 (888.225496)

USAGE EXAMPLES:
  mzrecal BSA.mzML
     Read BSA.mzML and BSA.mzid, write recalibration coefficents
     to BSA-recal.json.

  mzrecal -mzid BSA_comet.mzid -cal BSA_comet-recal.json BSA.mzML
     Read BSA.mzML and BSA_comet.mzid, write recalibration coefficents
     to BSA_comet-recal.json

  mzrecal -recal BSA.mzML
     Read BSA.mzML and BSA-recal.json, write recalibrated output to
     BSA-recal.mzML
```