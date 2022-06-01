
![CodeQL](https://github.com/524D/mzrecal/actions/workflows/codeql-analysis.yml/badge.svg)


# mzRecal

## What does mzRecal do?

mzrecal recalibrates mass spectrometry (MS1) data in mzML format, using peptide identifications in mzIdentML. mzRecal uses calibration functions based on the physics of the mass analyzer (FTICR, Orbitrap, TOF). The recalibration procedure was originally developed by Magnus Palmblad [[1]](#1)[[2]](#2). See also [msRecal](https://www.ms-utils.org/Taverna/msRecal.html) and [recal2](http://www.ms-utils.org/recal2.html) for more information on the predecessors of mzRecal.
Consuming and producing data in the same, open standard, format (mzML), mzRecal can be inserted into virtually any modular proteomics data analysis workflow, similar to msRecal [[3]](#3). This latest iteration of the software was described in an Application Note by Marissen and Palmblad in 2021 [[4]](#4).

Check section [Usage](#usage) for a more complete description.

## Running mzRecal

Ready-to-run executables of mzRecal for Linux and Microsoft Windows can be downloaded from <https://github.com/524D/mzrecal/releases/latest> (under "assets"). These executables have no external dependencies.

## Compiling

mzRecal is written in [Go](https://golang.org/). The software was tested with Go version 1.16

### Linux

On any recent Ubuntu/Debian, to install the prerequisites and download/build the executable:

* [install Go](https://golang.org/doc/install) using default install options

```bash
sudo apt install git
git clone https://github.com/524D/mzrecal
cd mzrecal; ./build.sh
```

The executables (both for Linux and for Windows)  are put in directory `~/tools`.

### Windows

On Windows, to install the prerequisites and download/build the executable:

* [install Go](https://golang.org/doc/install) using default install options
* [Install git](https://git-scm.com/download/win/) using default install options
* Restart Windows to add newly installed software to the PATH
* Open git bash (from the Windows start menu)
* Get mzRecal. From git bash prompt: `git clone https://github.com/524D/mzrecal`
* Build mzRecal. From git bash prompt: `cd mzrecal; ./build.sh`.
The executables (both for Windows and for Linux) are put in directory `tools` relative to the user's home directory.

## Input and output

mzRecal uses file formats specified by the Proteomics Standards Initiative
(PSI), notably [mzML](http://www.psidev.info/mzML) and [mzIdentML](http://www.psidev.info/mzidentml).

For recalibration, a peak-picked mzML file and corresponding
mzIdentML (file extension .mzid) file are needed as input.
Running `mzrecal` produces a recalibrated mzML file, plus a file with recalibration
parameters (.json format). The latter can be used to manually inspect
the calibration for each spectrum.

Note that the output mzML file will not contain the index wrapper
(which is optional according to the mzML specification, but still required by
some software). The [msconvert](http://proteowizard.sourceforge.net/download.html)
program from the ProteoWizard toolkit is recommended to add the index.

## Results

Recalibration affects the MS1 spectra as well as the precursor masses of the
MS2 spectra. Search engines commonly report the difference between theoretical
mass and measured mass for identified peptides. The following plot shows the
improvement of mzRecal on an Orbitrap and on a TOF dataset.
![ppm-histogram](./ppmerr.png)
This plot was made by [running plot-recal.R](./run-plot-recal.md) (included in the mzRecal repository) 

## Go packages for mzML and mzIdentML

The current version of the code embeds two internal Go packages, one for reading
mzIdentML and one for reading/writing mzML files. These packages will likely
be split into a separate module at a later time.

## <a name="usage"></a>Usage

The following is printed by running mzrecal -help

```text
USAGE:
  mzrecal [options] <mzMLfile>

  This program can be used to recalibrate MS data in an mzML file
  using peptide identifications in an accompanying mzID file.

OPTIONS:
  -acceptprofile
        Accept non-peak picked (profile) input.
        This is a kludge, and will be removed when mzRecal can perform peak-picking.
        By setting "acceptprofile", the value of option "calmult" is automatically
        set to 0 and the default of "minpeak" is set to 100000
  -cal filename
        filename for output of computed calibration parameters
  -calmult int
        only the topmost (<calmult> * <number of potential calibrants>)
        peaks are considered for computing the recalibration. <1 means all peaks. (default 10)
  -charge range
        charge range of calibrants, or the string "ident". If set to "ident",
        only the charge as found in the mzIdentMl file will be used for calibration. (default "1:5")
  -debug range
        Print debug output for given spectrum range e.g. 3:6
  -empty-non-calibrated
        Empty MS2 spectra for which the precursor was not recalibrated.
  -func function
        recalibration function to apply. If empty, a suitable
        function is determined from the instrument specified in the mzML file.
        Valid function names:
            FTICR, TOF, Orbitrap: Calibration function suitable for these instruments.
            POLY<N>: Polynomial with degree <N> (range 1:5)
            OFFSET: Constant m/z offset per spectrum.
  -mincals int
        minimum number of calibrants a spectrum should have to be recalibrated.
        If 0 (default), the minimum number of calibrants is set to the smallest number
        needed for the chosen recalibration function plus one. In any other case, if
        the specified number is too low for the calibration function, it is increased to
        the minimum needed value.
  -minpeak float
        minimum peak intensity to consider for computing the recalibration. (default 0)
  -mzid filename
        mzIdentMl filename
  -o filename
        filename of recalibrated mzML
  -ppmcal float
        0 (default): remove outlier calibrants according to HUPO-PSI mzQC,
           the rest is accepted.
        > 0: max mz error (ppm) for accepting a calibrant for calibration
  -ppmuncal float
        max mz error (ppm) for trying to use calibrant for calibration (default 10)
  -quiet
        Don't print any output except for errors
  -rt range
        rt window range(s) (default "-10.0:10.0")
  -scorefilter string
        filter for PSM scores to accept. Format:
        <CVterm1|scorename1>([<minscore1>]:[<maxscore1>])...
        When multiple score names/CV terms are specified, the first one on the list
        that matches a score in the input file will be used.
        The default contains reasonable values for some common search engines
        and post-search scoring software:
          MS:1002257 (Comet:expectation value)
          MS:1001330 (X!Tandem:expectation value)
          MS:1001159 (SEQUEST:expectation value)
          MS:1002466 (PeptideShaker PSM score)
          (default "MS:1002257(0.0:1e-2)MS:1001330(0.0:1e-2)MS:1001159(0.0:1e-2)MS:1002466(0.99:)")
  -specfilter range
        range of spectrum indices to calibrate (e.g. 1000:2000).
        Default is all spectra
  -stage int
        0 (default): do all calibration stages in one run
        1: only compute recalibration parameters
        2: perform recalibration using previously computer parameters
  -verbose
        Print more verbose progress information
  -version
        Show software version

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

ENVIRONMENT VARIABLES:
    When environment variable MZRECAL_DEBUG=1, extra information is added to the
    JSON file that can help checking the performance of mzrecal. 

USAGE EXAMPLES:
  mzrecal yeast.mzML
    Recalibrate yeast.mzML using identifications in yeast.mzid, write recalibrated
    result to yeast-recal.mzML and write recalibration coefficients yeast-recal.json.
    Default parameters are used. 

  mzrecal -ppmuncal 20 -scorefilter 'MS:1002257(0.0:0.001)' yeast.mzML
    Idem, but accept peptides with 20 ppm mass error and Comet expectation value <0.001
    as potential calibrants

NOTES:
    The mzML file that is produced after recalibration does not contain an index. If an
    index is required, we recommend post-processing the output file with msconvert
    (http://proteowizard.sourceforge.net/download.html).
```


## Acknowledgements

The authors gratefully acknowledge prior contributions from co-authors and collaborators in the development and testing of prior installments of the software.

## Funding

mzRecal was made possible in part due to funding from the ELIXIR Implementation Study "Crowd-sourcing the annotation of public proteomics datasets to improve data reusability".

## References

<a id="1">[1]</a> Palmblad M, Bindschedler LV, Gibson TM, Cramer R (2006).
Automatic internal calibration in liquid chromatography/Fourier transform ion cyclotron resonance mass spectrometry of protein digests. 
*Rapid Commun. Mass Spectrom.* 2006;20(20):3076-80.

<a id="2">[2]</a> Palmblad M, van der Burgt YEM, Dalebout H, Derks RJE, Schoenmaker B, Deelder AM (2009).
Improving mass measurement accuracy in mass spectrometry based proteomics by combining open source tools for chromatographic alignment and internal calibration.
*J. Proteomics.* 2009;72(4):722-4.

<a id="3">[3]</a> de Bruin JS, Deelder AM, Palmblad M (2012).
Scientific workflow management in proteomics.
*Mol. Cell. Proteomics.* 2012 Jul;11(7):M111.010595.

<a id="4">[4]</a> Marissen R, Palmblad M (2021).
mzRecal: universal MS1 recalibration in mzML using identified peptides in mzIdentML as internal calibrants.
*Bioinformatics.* 2021 Feb 4;btab056.
