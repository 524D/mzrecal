// Copyright 2018,2019,2020 Rob Marissen.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"C"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/524D/mzrecal/internal/mzidentml"
	"github.com/524D/mzrecal/internal/mzml"
)

// Program name and version, appended to software list in mzML output
const progName = "mzRecal"

var progVersion = `Unknown
Please build this program with script 'build.sh' so that the git version is shown here.`

// Format of output, if it ever changes we should still be able to parse
// output from old versions
const outputFormatVersion = "1.0"

// Peptides m/z values within mergeMzTol are merged
const mergeMzTol = float64(1e-7)
const massProton = float64(1.007276466879)
const massH2O = float64(18.0105647)

// CV parameters names
const cvParamSelectedIonMz = `MS:1000744`
const cvIsolationWindowTargetMz = `MS:1000827`
const cvFTICRSpectrometer = `MS:1000079`
const cvTOFSpectrometer = `MS:1000084`
const cvOrbiTrapSpectrometer = `MS:1000484`

// The calibration types that we can handle
// WARNING: This must be consistent with C enum calib_method_t in gsl_wrapper.go
const (
	calibNone = iota
	calibFTICR
	calibTOF
	calibOrbitrap
	calibOffset
	calibPoly1
	calibPoly2
	calibPoly3
	calibPoly4
	calibPoly5
)

// Command line parameters
type params struct {
	stage              *int // Compute recal parmeters (1), recalibrate (2) or both (0)
	mzMLFilename       *string
	mzMLRecalFilename  *string
	mzIdentMlFilename  *string
	calFilename        *string  // Filename where JSON calibration parameters will be written
	emptyNonCalibrated *bool    // Empty MS2 spectra for which the precursor was not recalibrated
	minCal             *int     // minimum number of calibrants a spectrum should have to be recalibrated
	minPeak            *float64 // minimum intensity of peaks to be considered for recalibrating
	calPeaks           *int     // number of peaks per potential calibrant to consider
	rtWindow           *string  // retention time window
	lowRT              float64  // lower rt window boundary
	upRT               float64  // upper rt window boundary
	mzErrPPM           *float64 // max mz error for trying a calibrant in calibrant
	mzTargetPPM        *float64 // max mz error for accepting a calibrant in calibration
	recalMethod        *string  // Recal method as specfied by user
	scoreFilter        *string  // PSM score filter to apply
	charge             *string  // Charge range for calibrants
	useIdentCharge     bool     // Use only chage as found in identification
	minCharge          int      // min charge for calibrants
	maxCharge          int      // max charge for calibrants
	args               []string // Addtional values passed on the command line
	debug              bool     // Enable debug info (environment variable MZRECAL_DEBUG=1)
}

// Calibrant as read from mzid file (or build in), with uncharged mass
type identifiedCalibrant struct {
	name          string
	mass          float64 // Uncharged mass
	retentionTime float64
	idCharge      int  // Charge state at identification
	singleCharged bool // true if only charge state 1 should be considered
}

// m/z value for calibrant
type chargedCalibrant struct {
	idCal  *identifiedCalibrant
	charge int     // assumed charge for finding m/z peak
	mz     float64 // m/z value, computed from uncharged mass and charge
}

// Calibrants with same m/z
type calibrant struct {
	chargedCals []chargedCalibrant
	mz          float64 // computed mz of the calibrant (copy of chargedCals[0])
	mzMeasured  float64 // mz of the best candidate peak
}

// recalParams contains recalibration parameters for each spectrum,
// in addition to generic recalibration data for the whole file
type recalParams struct {
	// Version of recalibration parameters, used when storing/loading
	// parameters in JSON format for different version of the software
	MzRecalVersion string
	RecalMethod    string // Recalibration method used (TOF/FTICR/Orbitrap)
	SpecRecalPar   []specRecalParams
}

type specDebugInfo struct {
	CalsInRTWindow   int
	CalsInMassWindow int
	CalsUsed         int
	TotalIonCurrent  float64 `json:",omitempty"`
	IonInjectionTime float64 `json:",omitempty"`
}

// specRecalParams contain the recalibration parameters for each
// spectrum. RecalMethod (from type recalParams) determines which
// computation must be done with these parameters to obtain the
// final calibration
type specRecalParams struct {
	SpecIndex int
	P         []float64
	DebugInfo []specDebugInfo `json:",omitempty"`
}

type scoreRange struct {
	minScore float64 // Minimum score to accept
	maxScore float64 // Maximum score to accept
	priority int     // Priority of the score, lowest is best
}

type scoreFilter map[string]scoreRange

type mzRange struct {
	min float64
	max float64
}

var fixedCalibrants = []identifiedCalibrant{

	// cyclosiloxanes, H6nC2nOnSin
	identifiedCalibrant{
		name:          `cyclosiloxane6`,
		mass:          444.1127481,
		retentionTime: -math.MaxFloat64, // Indicates any retention time
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane7`,
		mass:          518.1315394,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane8`,
		mass:          592.1503308,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane9`,
		mass:          666.1691221,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane10`,
		mass:          740.1879134,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane11`,
		mass:          814.2067048,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
	identifiedCalibrant{
		name:          `cyclosiloxane12`,
		mass:          888.2254961,
		retentionTime: -math.MaxFloat64,
		idCharge:      1,
		singleCharged: true,
	},
}

// Masses of amino acids (minus H2O)
var aaMass = map[rune]float64{
	'A': 71.0371138,
	'C': 103.0091848,
	'D': 115.0269430,
	'E': 129.0425931,
	'F': 147.0684139,
	'G': 57.0214637,
	'H': 137.0589119,
	'I': 113.0840640,
	'K': 128.0949630,
	'L': 113.0840640,
	'M': 131.0404849,
	'N': 114.0429274,
	'P': 97.0527638,
	'O': 237.1477269, // Pyrrolysine
	'Q': 128.0585775,
	'R': 156.1011110,
	'S': 87.0320284,
	'T': 101.0476785,
	'U': 144.9595902, // Selenocysteine
	'V': 99.0684139,
	'W': 186.0793129,
	'Y': 163.0633285,
}

// Parse string like "-12:6" into 2 values, -12 and 6
// Parameters min and max are the "default" min/max values,
// when a value is not specified (e.g. "-12:"), the defauls is assigned
func parseIntRange(r string, min int, max int) (int, int, error) {
	re := regexp.MustCompile(`\s*(\-?\d*):(\-?\d*)`)
	m := re.FindStringSubmatch(r)
	minOut := min
	maxOut := max
	if len(m) >= 2 && m[1] != "" {
		minOut, _ = strconv.Atoi(m[1])
		if minOut < min {
			minOut = min
		}
	}
	if len(m) >= 3 && m[2] != "" {
		maxOut, _ = strconv.Atoi(m[2])
		if maxOut > max {
			maxOut = max
		}
	}
	var err error
	if minOut > maxOut {
		err = errors.New("parseIntRange min>max")
		minOut = maxOut
	}
	return minOut, maxOut, err
}

// Parse string like "-12.01e1:+6" into 2 values, -120.1 and 6.0
// Parameters min and max are the "default" min/max values,
// when a value is not specified (e.g. "-12.01e1:"), the defauls is assigned
func parseFloat64Range(r string, min float64, max float64) (
	float64, float64, error) {
	re := regexp.MustCompile(`\s*([-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?):([-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?)`)
	m := re.FindStringSubmatch(r)
	minOut := min
	maxOut := max
	if len(m) >= 2 && m[1] != "" {
		minOut, _ = strconv.ParseFloat(m[1], 64)
		if minOut < min {
			minOut = min
		}
	}
	if len(m) >= 4 && m[3] != "" {
		maxOut, _ = strconv.ParseFloat(m[3], 64)
		if maxOut > max {
			maxOut = max
		}
	}
	var err error
	if minOut > maxOut {
		err = errors.New("parseFloat64Range min>max")
		minOut = maxOut
	}
	return minOut, maxOut, err
}

// Compute the lowest isotope mass of the peptide
func pepMass(pepSeq string) (float64, error) {
	m := massH2O
	for _, aa := range pepSeq {
		aam, ok := aaMass[aa]
		if !ok {
			return 0.0, errors.New("Invalid amino acid")
		}
		m += aam
	}
	return m, nil
}

// This function creates a slice with potential calibrants
// Calibrants are obtained from 2 sources:
// - Identied peptides (from mzid file)
// - Build-in list of fixed calibrants (cyclosiloxanes)
// Identified peptides are only used if they pass the score filter
// For each calibrant, it:
// - computes the mass of the lightest isotope
// - get the retention name, retentionTime, spectrum
func makeCalibrantList(mzIdentML *mzidentml.MzIdentML, scoreFilt scoreFilter,
	par params) ([]identifiedCalibrant, error) {
	// Create slice for the number of calibrants that we expect to have
	cals := make([]identifiedCalibrant, 0, mzIdentML.NumIdents()+len(fixedCalibrants))
	for i := 0; i < mzIdentML.NumIdents(); i++ {
		ident, err := mzIdentML.Ident(i)
		if err != nil {
			return nil, err
		}
		//		log.Printf("indent %+v\n", ident)
		scoreOK := false
		curPrio := math.MaxInt32
		for _, cv := range ident.Cv {
			// Check if the CV accession number or CV name matches scorefilter
			filt, ok := scoreFilt[cv.Accession]
			if !ok {
				filt, ok = scoreFilt[cv.Name]
			}
			if ok {
				if filt.priority < curPrio {
					var score float64
					score, err = strconv.ParseFloat(cv.Value, 64)
					if err != nil {
						return nil, errors.New("Invalid score value " + cv.Value)
					}
					scoreOK = score >= filt.minScore && score <= filt.maxScore
				}
			}
		}
		if scoreOK {
			var cal identifiedCalibrant
			m, err := pepMass(ident.PepSeq)
			if err == nil { // Skip if mass cannot be computed
				cal.name = ident.PepID
				cal.retentionTime = ident.RetentionTime
				cal.idCharge = ident.Charge
				cal.singleCharged = false
				cal.mass = m + ident.ModMass
				cals = append(cals, cal)
			}
		} else {
			//			log.Print(ident.PepID + " does not match score filter.")
		}
	}
	//	log.Print(len(cals), " of ", mzIdentML.NumIdents(), " identifications usable for calibration.")
	if len(cals) == 0 {
		log.Print("No identified spectra will be used as calibrant. Is the specified scorefilter applicable for this file?")
	}
	cals = append(cals, fixedCalibrants...)
	sort.Slice(cals,
		func(i, j int) bool { return cals[i].retentionTime < cals[j].retentionTime })

	return cals, nil
}

func calibsInRtWindows(rtMin, rtMax float64, allCals []identifiedCalibrant) ([]identifiedCalibrant, error) {

	// Find the indexes of the calibrants within the retention time window
	i1 := sort.Search(len(allCals), func(i int) bool { return allCals[i].retentionTime >= rtMin })
	i2 := sort.Search(len(allCals), func(i int) bool { return allCals[i].retentionTime > rtMax })

	// Find calibrants that elute at all retention times
	// These have elution time -math.MaxFloat64 and are located at the
	// start of the list of calibrants. Thus, to search them, we simply
	// search from the start for calibrants with elution time -math.MaxFloat64
	var i3 int
	for i3 = 0; i3 < len(allCals) && allCals[i3].retentionTime == -math.MaxFloat64; i3++ {
	}

	var cals = make([]identifiedCalibrant, 0, (i2-i1)+i3)
	cals = append(cals, allCals[i1:i2]...)
	cals = append(cals, allCals[0:i3]...)

	return cals, nil
}

// mergeSameMzCals merges all calibrants that have the same m/z or
// nearly the same m/z. The m/z of the first calibrant that was encountered
// is retained. The list of calibrants with their chargestate is appended
// to the final list of calibrants.
func mergeSameMzCals(chargedCalibrants []chargedCalibrant) []calibrant {
	mcals := make([]calibrant, 0, len(chargedCalibrants))
	// sort calibrants by mass
	sort.Slice(chargedCalibrants,
		func(i, j int) bool { return chargedCalibrants[i].mz < chargedCalibrants[j].mz })

	prevMz := float64(-1)
	for _, cal := range chargedCalibrants {
		if math.Abs(cal.mz-prevMz) < mergeMzTol {
			mcals[len(mcals)-1].chargedCals = append(mcals[len(mcals)-1].chargedCals, cal)
		} else {
			var newCal calibrant
			newCal.chargedCals = make([]chargedCalibrant, 1)
			newCal.chargedCals[0] = cal
			newCal.mz = cal.mz
			mcals = append(mcals, newCal)
			prevMz = cal.mz
		}
	}
	return mcals
}

func newChargedCalibrant(charge int, idCal *identifiedCalibrant) chargedCalibrant {
	var chargedCal chargedCalibrant

	fCharge := float64(charge)
	chargedCal.mz = (idCal.mass + fCharge*massProton) / fCharge
	chargedCal.idCal = idCal
	chargedCal.charge = charge
	return chargedCal
}

func instrument2RecalMethod(mzML *mzml.MzML) (int, string, error) {
	instruments, err := mzML.MSInstruments()
	if err != nil {
		return 0, ``, err
	}
	for _, instr := range instruments {
		switch instr {
		case cvFTICRSpectrometer:
			return calibFTICR, `FTICR`, nil
		case cvTOFSpectrometer:
			return calibTOF, `TOF`, nil
		case cvOrbiTrapSpectrometer:
			return calibOrbitrap, `Orbitrap`, nil
		}
	}
	// FIXME: Implement other instruments
	log.Println("WARNING: No recalibration method for instrument, using POLY2 recalibration")
	return calibPoly2, `POLY2`, nil
}

func recalMethodStr2Int(recalMethodStr string) (int, error) {
	var recalMethod int
	switch strings.ToUpper(recalMethodStr) {
	case `FTICR`:
		recalMethod = calibFTICR
	case `TOF`:
		recalMethod = calibTOF
	case `ORBITRAP`:
		recalMethod = calibOrbitrap
	case `OFFSET`:
		recalMethod = calibOffset
	case `POLY1`:
		recalMethod = calibPoly1
	case `POLY2`:
		recalMethod = calibPoly2
	case `POLY3`:
		recalMethod = calibPoly3
	case `POLY4`:
		recalMethod = calibPoly4
	case `POLY5`:
		recalMethod = calibPoly5
	default:
		return 0, errors.New("Unknown recalibration method: " + recalMethodStr)
	}
	return recalMethod, nil
}

// Get the minimum and maximum mz in a slice of peaks
// Potentially, these values could be obtained from the corresponding
// tags in the mzML file, be we don't want te depend on that.
func mzRangePeaks(peaks []mzml.Peak) mzRange {
	var r mzRange

	if len(peaks) > 0 {
		r.min = peaks[0].Mz
		r.max = peaks[0].Mz
		for _, p := range peaks {
			m := p.Mz
			if m < r.min {
				r.min = m
			}
			if m > r.max {
				r.max = m
			}
		}
	}
	return r
}

// Remove calibrants that are outside a given mzrange
// We modify the slice of calibrants in place, hence it is
// passed as a pointer
func filterMzCalibs(calibrants *[]calibrant, r mzRange) {
	if calibrants != nil {
		k := int(0) // Index of calibrants that we want to keep
		for i, c := range *calibrants {
			if c.mz >= r.min && c.mz <= r.max {
				// Calibrant is mz range
				// If calibrant is not in range, k is not incremented,
				// so it will be removed/overwritten
				// Optimization: only copy is source and destination are the different
				if k < i {
					(*calibrants)[k] = (*calibrants)[i]
				}
				k++
			}
		}
		*calibrants = (*calibrants)[:k] // Change slice length
	}
}

func computeRecal(mzML *mzml.MzML, idCals []identifiedCalibrant, par params) (recalParams, error) {
	var recal recalParams
	var err error
	var recalMethod int

	recal.MzRecalVersion = outputFormatVersion
	if *par.recalMethod == `` {
		recalMethod, recal.RecalMethod, err = instrument2RecalMethod(mzML)
	} else {
		recalMethod, err = recalMethodStr2Int(*par.recalMethod)
		recal.RecalMethod = *par.recalMethod
	}
	if err != nil {
		return recal, err
	}
	// Update the minimum number of calibrants
	// according to the calibration method
	nrCalPars := getNrCalPars(recalMethod)
	if *par.minCal == 0 {
		*par.minCal = nrCalPars + 1
	} else {
		if *par.minCal < nrCalPars {
			*par.minCal = nrCalPars
		}
	}

	numSpecs := mzML.NumSpecs()

	for i := 0; i < numSpecs; i++ {
		centroid, err := mzML.Centroid(i)
		if err != nil {
			return recal, err
		}
		if !centroid {
			return recal, errors.New(`input mzML file must contain centroid data, not profile data`)
		}
		msLevel, err := mzML.MSLevel(i)
		if err != nil {
			return recal, err
		}
		if msLevel == 1 {
			retentionTime, err := mzML.RetentionTime(i)
			if err != nil {
				return recal, err
			}
			specCals, err := calibsInRtWindows(retentionTime+par.lowRT,
				retentionTime+par.upRT, idCals)
			if err != nil {
				return recal, err
			}

			// Make slice with mz values for all calibrants
			// For efficiency, pre-allocate (more than) enough elements
			chargedCalibrants := make([]chargedCalibrant, 0,
				len(specCals)*(par.maxCharge-par.minCharge+1))
			for j, cal := range specCals {
				//				log.Printf("Calibrating spec %d, rt %f, calibrants: %+v\n", i, retentionTime, cal)
				if cal.singleCharged {
					chargedCalibrants = append(chargedCalibrants, newChargedCalibrant(1, &specCals[j]))
				} else {
					if par.useIdentCharge {
						chargedCalibrants = append(chargedCalibrants, newChargedCalibrant(cal.idCharge, &specCals[j]))
					} else {
						for charge := par.minCharge; charge <= par.maxCharge; charge++ {
							chargedCalibrants = append(chargedCalibrants, newChargedCalibrant(charge, &specCals[j]))
						}
					}
				}
			}
			calibrants := mergeSameMzCals(chargedCalibrants)

			peaks, err := mzML.ReadScan(i)
			if err != nil {
				log.Fatalf("computeRecal ReadScan failed for spectrum %d: %v",
					i, err)
			}
			// Remove calibrants outside of measured range
			r := mzRangePeaks(peaks)
			filterMzCalibs(&calibrants, r)

			matchingCals := calibrantsMatchPeaks(peaks, calibrants, par)

			specRecalPar, calibrantsUsed, err := recalibrateSpec(i, recalMethod,
				matchingCals, par)
			if err != nil {
				log.Printf("computeRecal calibration failed for spectrum %d: %v",
					i, err)
			}
			if par.debug {
				debugInfo := make([]specDebugInfo, 1, 1)
				debugInfo[0].CalsInRTWindow = len(calibrants)
				debugInfo[0].CalsInMassWindow = len(matchingCals)
				debugInfo[0].CalsUsed = len(calibrantsUsed)
				iit, err := mzML.IonInjectionTime(i)
				if err != nil {
					log.Printf("computeRecal IonInjectionTime failed for spectrum %d: %v",
						i, err)
				}
				if !math.IsNaN(iit) {
					debugInfo[0].IonInjectionTime = iit
				}
				tic, _ := mzML.TotalIonCurrent(i)
				if !math.IsNaN(tic) {
					debugInfo[0].TotalIonCurrent = tic
				}
				specRecalPar.DebugInfo = debugInfo
			}
			recal.SpecRecalPar = append(recal.SpecRecalPar, specRecalPar)

			debugLogSpecs(i, numSpecs, retentionTime, peaks, matchingCals, par,
				calibrantsUsed, recalMethod, specRecalPar)
			debugRegisterCalUsed(i, matchingCals, par, calibrantsUsed)

			// log.Printf("Spec %d retention match %d mz match %d calib used %d",
			// 	i, len(mzCalibrants), len(mzMatchingCals), calibrantsUsed)
		}
	}
	debugListUnusedCalibrants(idCals)
	return recal, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Return a slice with only the peaks that we want to base the calibration on
// Only the most intence peaks are used, filtered by number of potential calibrants
// and absolute peak size
func filterPeaks(peaks []mzml.Peak, par params, c int) []mzml.Peak {
	// Make a copy of the slice, we don't want to sort the origial
	peaksNew := make([]mzml.Peak, len(peaks))
	copy(peaksNew, peaks)
	// sort by intensity, so the most intense peaks are at the front
	sort.Slice(peaksNew,
		func(i, j int) bool { return peaksNew[i].Intens > peaksNew[j].Intens })
	// Should we filter on number of potential calibrants?
	if *par.calPeaks > 0 {
		// The number of peaks to consider is:
		// calPeaks * the number of potential calibrants
		n1 := *par.calPeaks * c
		// Set slice length to desired number of peaks,
		// or length of original if that is less
		peaksNew = peaksNew[:min(n1, len(peaksNew))]
	}
	// Should we filter on absolute peak size?
	if *par.minPeak > 0 {
		// Find the lowest peak that we still have to consider
		n2 := sort.Search(len(peaksNew),
			func(i int) bool { return peaksNew[i].Intens < *par.minPeak })
		peaksNew = peaksNew[:min(n2, len(peaksNew))]
	}
	return peaksNew
}

func calibrantsMatchPeaks(peaks []mzml.Peak, calibrants []calibrant,
	par params) []calibrant {
	matchingCals := make([]calibrant, 0, len(calibrants))

	// Remove the peaks that are too small
	peaks = filterPeaks(peaks, par, len(calibrants))

	// Sort peaks by mass, so we can find matching masses quickly
	sort.Slice(peaks,
		func(i, j int) bool { return peaks[i].Mz < peaks[j].Mz })

	// For each potential calibrant, find highest peak within mz window
	for _, calibrant := range calibrants {
		mz := calibrant.mz
		mzErr := *par.mzErrPPM * mz / 1000000.0
		peak := maxPeakInMzWindow(mz-mzErr, mz+mzErr, peaks)
		// If a peak was found
		if peak.Intens != 0 {
			calibrant.mzMeasured = peak.Mz
			matchingCals = append(matchingCals, calibrant)
		}
	}
	return matchingCals
}

// maxPeakInMzWindow returns the highest intensity peak in a given mz window.
// Peaks must be ordered by mz prior to calling this function
// If no peak was found, peak.intensity will be 0
func maxPeakInMzWindow(mzMin, mzMax float64, peaks []mzml.Peak) mzml.Peak {

	// Find the indexes of the calibrants within the retention time window
	i1 := sort.Search(len(peaks), func(i int) bool { return peaks[i].Mz >= mzMin })
	i2 := sort.Search(len(peaks), func(i int) bool { return peaks[i].Mz > mzMax })

	var peak mzml.Peak // auto initialzed to 0.0, 0.0
	for i := i1; i < i2; i++ {
		if peaks[i].Intens > peak.Intens {
			peak = peaks[i]
		}
	}
	return peak
}

func writeRecal(recal recalParams, par params) error {
	f, err := os.Create(*par.calFilename)
	if err != nil {
		return err
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent(``, `  `) // Make output easier to read for humans
	e.Encode(recal)
	return nil
}

func readRecal(par params) (recalParams, error) {
	var recal recalParams
	f, err := os.Open(*par.calFilename)
	if err != nil {
		return recal, err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	err = d.Decode(&recal)
	return recal, err
}

func writeRecalMzML(mzML mzml.MzML, recal recalParams, par params) error {
	_ = recal
	_ = par
	f, err := os.Create(*par.mzMLRecalFilename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = mzML.Write(f)
	return err
}

func parseScoreFilter(scoreFilterStr string) (scoreFilter, error) {
	scoreFilt := make(scoreFilter)

	re := regexp.MustCompile(`([^\(]+)\(([^\)]*)\)`)
	matchedStringsList := re.FindAllStringSubmatch(scoreFilterStr, -1)
	if matchedStringsList != nil {
		for n, matchedStrings := range matchedStringsList {

			scoreName := matchedStrings[1]
			scoreRangeStr := matchedStrings[2]
			_, ok := scoreFilt[scoreName]
			if ok {
				return nil, errors.New(scoreName + ` defined more than once.`)
			}
			minScore, maxScore, err := parseFloat64Range(scoreRangeStr,
				-math.MaxFloat64, math.MaxFloat64)

			if err != nil {
				return nil, errors.New(`Invalid range for score ` + scoreName)
			}
			scRange := scoreRange{minScore: minScore, maxScore: maxScore, priority: n}
			scoreFilt[scoreName] = scRange
		}
	}

	return scoreFilt, nil
}

type rtSpec struct {
	rt   float64
	spec int
}

type rtSpecs []rtSpec

func (a rtSpecs) Len() int           { return len(a) }
func (a rtSpecs) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a rtSpecs) Less(i, j int) bool { return a[i].rt < a[j].rt }

// Find the index of the MS1 scan that has a retention time just less than
// the retention time in rt
func findRtMs1(rt float64, rtOfSpecs rtSpecs) int {
	j := sort.Search(len(rtOfSpecs), func(i int) bool { return rtOfSpecs[i].rt >= rt })
	if j > 0 {
		j--
	}
	return rtOfSpecs[j].spec
}

// initRtMs1 created a data structure needed by findRtMs1
func initRtMs1(mzML mzml.MzML) (rtSpecs, error) {
	var rtOfSpec rtSpec
	numSpecs := mzML.NumSpecs()
	rtOfMs1Specs := make(rtSpecs, 0, numSpecs)

	for i := 0; i < numSpecs; i++ {
		// Get retention time of MS1 spectra
		MSLevel, err := mzML.MSLevel(i)
		if err != nil {
			return nil, err
		}
		if MSLevel == 1 {
			rtOfSpec.spec = i
			rtOfSpec.rt, err = mzML.RetentionTime(i)
			if err != nil {
				return nil, err
			}
			rtOfMs1Specs = append(rtOfMs1Specs, rtOfSpec)
		}
	}
	if len(rtOfMs1Specs) == 0 {
		return nil, fmt.Errorf("No MS1 spectra found, calibration not possible")
	}
	sort.Sort(rtOfMs1Specs)

	return rtOfMs1Specs, nil
}

func updatePrecursorMz(mzML mzml.MzML, recal recalParams, par params) error {

	var precursorsUpdated, precursorsTotal int
	recalMethod, err := recalMethodStr2Int(recal.RecalMethod)
	if err != nil {
		return err
	}

	// Make map to lookup recal parameters for a given spectrum index
	specIndex2recalIndex := make(map[int]int)
	for i, specRecalPar := range recal.SpecRecalPar {
		specIndex2recalIndex[specRecalPar.SpecIndex] = i
	}

	rtOfMs1Specs, err := initRtMs1(mzML)
	if err != nil {
		return err
	}
	numSpecs := mzML.NumSpecs()
	for i := 0; i < numSpecs; i++ {
		// Only update precursors for MS2
		MSLevel, err := mzML.MSLevel(i)
		if err != nil {
			return err
		}
		if MSLevel == 2 {
			precursorsTotal++
			// The precursor MS1 spectrum is the one for which we have recalibration
			// Find the MS1 spctrum that belongs to this MS2, so that
			// we can recalibrate the precursor mass of the MS2.
			// We cannot use SpectrumRef to obtain the parent spectrum
			// because it is not always present (i.e. SCIEX)
			// therefore, we assume that previous (retention time wise) MS1 spectrum
			// is the correct one.
			rt, err := mzML.RetentionTime(i)
			if err != nil {
				return err
			}
			ms1ScanIndex := findRtMs1(rt, rtOfMs1Specs)

			precursors, err := mzML.GetPrecursors(i)
			if err != nil {
				return err
			}
			for _, precursor := range precursors {
				recalIndex, ok := specIndex2recalIndex[ms1ScanIndex]
				if !ok {
					log.Printf("Recalibration parameters missing for scanIndex %d)",
						ms1ScanIndex)
				}
				if ok && recal.SpecRecalPar[recalIndex].P != nil {
					specRecalPar := recal.SpecRecalPar[recalIndex]
					recalPars := setRecalPars(recalMethod, specRecalPar)
					recalIsolationWindow(&precursor, &recalPars, par, i)
					if recalSelectedIons(&precursor, &recalPars, par, i, numSpecs) {
						precursorsUpdated++
					}
				} else {
					if *par.emptyNonCalibrated {
						// Empty the spectrum
						var peaks []mzml.Peak
						mzML.UpdateScan(i, peaks, true, true)
					}
				}
			}
		}
	}
	log.Printf("MS2 count: %d Updated:%d\n", precursorsTotal, precursorsUpdated)
	return nil
}

func recalIsolationWindow(precursor *mzml.XMLprecursor, recalPars *CalParams, par params, specNr int) {
	isolationWindow := precursor.IsolationWindow
	for k, cvParam := range isolationWindow.CvParam {
		if cvParam.Accession == cvIsolationWindowTargetMz {
			mz, err := strconv.ParseFloat(cvParam.Value, 64)
			if err != nil {
				log.Printf("Invalid mz value %s (spec %d)",
					cvParam.Value, specNr)
			} else {
				mzNew := mzRecal(mz, recalPars)
				isolationWindow.CvParam[k].Value =
					strconv.FormatFloat(mzNew, 'f', 8, 64)
				break
			}
		}
	}
}

func recalSelectedIons(precursor *mzml.XMLprecursor, recalPars *CalParams,
	par params, specNr int, numSpecs int) bool {
	var updated bool
	for _, selectedIon := range precursor.SelectedIonList.SelectedIon {
		for k, cvParam := range selectedIon.CvParam {
			if cvParam.Accession == cvParamSelectedIonMz {
				mz, err := strconv.ParseFloat(cvParam.Value, 64)
				if err != nil {
					log.Printf("Invalid mz value %s (spec %d)",
						cvParam.Value, specNr)
				} else {
					mzNew := mzRecal(mz, recalPars)
					selectedIon.CvParam[k].Value =
						strconv.FormatFloat(mzNew, 'f', 8, 64)
					debugLogPrecursorUpdate(specNr, numSpecs, mz, mzNew, par)
					updated = true
					break
				}
			}
		}
	}
	return updated
}

// doRecal glues together all the steps to produce a
// re-calibrated mzML file:
// Read mzML file
// Read racalibration parameters from JSON file
// Recalibrate each spectrum
// Add our program name and version to the mlML software list
// Write recalibrated mlML file
func doRecal(par params) {
	mzFile, err := os.Open(*par.mzMLFilename)
	if err != nil {
		log.Fatalf("Open %s: mzMLfile %v", *par.mzMLFilename, err)
	}
	defer mzFile.Close()
	mzML, err := mzml.Read(mzFile)
	if err != nil {
		log.Fatalf("mzml.Read: error return %v", err)
	}

	recal, err := readRecal(par)
	if err != nil {
		log.Fatalf("readRecal: error return %v", err)
	}

	calibMzML(par, mzML, recal)
}

// calibMzML re-calibrates an mzML file:
// Recalibrate each spectrum
// Add our program name and version to the mlML software list
// Write recalibrated mlML file
func calibMzML(par params, mzML mzml.MzML, recal recalParams) {
	recalMethod, err := recalMethodStr2Int(recal.RecalMethod)

	for _, specRecalPar := range recal.SpecRecalPar {
		// Skip spectra for which no recalibration coefficients are available
		if specRecalPar.P != nil {
			specIndex := specRecalPar.SpecIndex
			recalPars := setRecalPars(recalMethod, specRecalPar)
			peaks, err1 := mzML.ReadScan(specIndex)
			if err1 != nil {
				log.Fatalf("readRecal: mzML.ReadScan %v", err1)
			}
			for i, peak := range peaks {
				mzNew := mzRecal(peak.Mz, &recalPars)
				peaks[i].Mz = mzNew
			}
			mzML.UpdateScan(specIndex, peaks, true, false)
		}
	}

	mzML.AppendSoftwareInfo(progName, progVersion)

	err = updatePrecursorMz(mzML, recal, par)
	if err != nil {
		log.Fatalf("updatePrecursorMz: %v", err)
	}

	err = writeRecalMzML(mzML, recal, par)
	if err != nil {
		log.Fatalf("writeRecalMzML: error return %v", err)
	}
}

func makeRecalCoefficients(par params) (mzML mzml.MzML, recal recalParams) {
	scoreFilt, err := parseScoreFilter(*par.scoreFilter)
	if err != nil {
		log.Fatalf("Invalid parameter 'scoreFilter': %v", err)
	}

	f1, err := os.Open(*par.mzIdentMlFilename)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer f1.Close()
	mzIdentML, err := mzidentml.Read(f1)
	if err != nil {
		log.Fatalf("mzidentml.Read: error return %v", err)
	}
	idCals, err := makeCalibrantList(&mzIdentML, scoreFilt, par)
	if err != nil {
		log.Fatal("makeCalibrantList failed")
	}

	f2, err := os.Open(*par.mzMLFilename)
	if err != nil {
		log.Fatalf("Open: mzMLfile %v", err)
	}
	defer f2.Close()
	mzML, err = mzml.Read(f2)
	if err != nil {
		log.Fatalf("mzml.Read: error return %v", err)
	}

	recal, err = computeRecal(&mzML, idCals, par)
	if err != nil {
		log.Fatalf("computeRecal: error return %v", err)
	}

	err = writeRecal(recal, par)
	if err != nil {
		log.Fatalf("writeRecal: error return %v", err)
	}
	return mzML, recal
}

// sanatizeParams does some checks on parameters, and fills missing
// filenames if possible
func sanatizeParams(par *params) {
	progName := filepath.Base(os.Args[0])

	if len(par.args) != 1 {
		fmt.Fprintf(os.Stderr, `Last argument must be name of mzML file.
Type %s --help for usage
`, progName)
		os.Exit(2)
	}

	mzml := par.args[0]
	par.mzMLFilename = &mzml
	var extension = filepath.Ext(mzml)
	var startName = mzml[0 : len(mzml)-len(extension)]

	if *par.mzIdentMlFilename == "" {
		*par.mzIdentMlFilename = startName + ".mzid"
	}
	if *par.calFilename == "" {
		*par.calFilename = startName + "-recal.json"
	}
	if *par.mzMLRecalFilename == "" {
		*par.mzMLRecalFilename = startName + "-recal.mzML"
	}

	var err error
	par.lowRT, par.upRT, err = parseFloat64Range(*par.rtWindow,
		-math.MaxFloat64, math.MaxFloat64)
	if err != nil {
		fmt.Fprintf(os.Stderr, `Invalid rtWindow.
Type %s --help for usage
`, progName)
		os.Exit(2)
	}
	if *par.charge == `ident` {
		par.useIdentCharge = true
	} else {
		par.minCharge, par.maxCharge, err = parseIntRange(*par.charge,
			1, 5)
		if err != nil {
			fmt.Fprintf(os.Stderr, `Invalid charge range.
	Type %s --help for usage
	`, progName)
			os.Exit(2)
		}
	}
}

func usage() {
	progName := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr,
		`USAGE:
  %s [options] <mzMLfile>

  This program can be used to recalibrate MS data in an mzML file
  using peptide identifications in an accompanying mzID file.

OPTIONS:
`, progName)
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr,
		`
BUILD-IN CALIBRANTS:
  In addition to the identified peptides, %s will also use
  for recalibration a number of compounds that are commonly found in many
  samples. These compound are all assumed to have +1 charge. The following
  list shows the build-in compounds with their (uncharged) masses:
`, progName)

	for _, cal := range fixedCalibrants {
		fmt.Fprintf(os.Stderr, "     %s (%f)\n", cal.name, cal.mass)
	}

	fmt.Fprintf(os.Stderr,
		`
EXECUTION STAGES:
    Recalibration consists of 2 stages. By default they are executed consequtively,
    but it is also possible to execute them seperately by specifying the -stage flag:
    1) Computation of recalibration coefficients. The coefficients are stored
        in a JSON file.
        This stage reads an mzML file and mzID file, matches measured peaks to
        computed m/z values and computes recalibration coefficents using a method
        that is usefull for the instrument type. The instrument type (and other
        relavant values) are determined from the CV terms in the input files.
    2) Creating a recalibrated version of the MS file.
        This stage reads the mzML file and JSON file with recalibration values,
        computes recalibrated m/z values for all peaks in spectra for which
        a valid recalibration was found, and writes a recalibrated mzML file.

ENVIRONMENT VARIABLES:
	When environment variable MZRECAL_DEBUG=1, extra information is added to the
	JSON file that can help checking the performance of %s. 

USAGE EXAMPLES:
  %s BSA.mzML
    Recalibrate BSA.mzML using identifications in BSA.mzid, write recalibrated
    result to BSA-recal.mzML and write recalibration coefficients BSA-recal.json.
    Default parameters are used. 

  %s -ppmuncal 20 -ppmcal 1.5 BSA.mzML
    Idem, but accept peptides with 20 ppm as potential calibrants, after
    recalibration all accepted peptides must be within 1.5 ppm

  %s -calmult 20
    Idem, but the number of peaks that are considered for matching are the
    number of potential calibrants times 20
`, progName, progName, progName, progName)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var par params

	par.recalMethod = flag.String("func",
		"",
		`recalibration function to apply. If empty, a suitable
function is determined from the instrument specified in the mzML file.
Valid function names:
    FTICR, TOF, Orbitrap: Calibration function suitable for these instruments.
    POLY<N>: Polynomial with degee <N> (range 1:5)
    OFFSET: Constant m/z offset per spectrum.`)
	par.stage = flag.Int("stage", 0,
		`0: do all calibration stages in one run
1: only compute recalibration parameters
2: perform recalibration using previously computer parameters
NOTE: The mzML file that is produced after recalibration does not contain an
      index. If an index is required, we recommend post-processing the output file 
      with msconvert (http://proteowizard.sourceforge.net/download.html).`)
	par.mzMLRecalFilename = flag.String("o",
		"",
		`filename of recalibrated mzML`)
	par.mzIdentMlFilename = flag.String("mzid",
		"",
		`mzIdentMl filename`)
	par.calFilename = flag.String("cal",
		"",
		`filename for output of computed calibration parameters`)
	par.emptyNonCalibrated = flag.Bool("empty-non-calibrated", false,
		`Empty MS2 spectra for which the precursor was not recalibrated.`)
	par.minCal = flag.Int("mincals",
		0,
		`minimum number of calibrants a spectrum should have to be recalibrated.
If 0, the minimum number of calibrants is set to the smallest number needed
for the choosen recalibration function plus one. In any other case, is the
specified number is too low for the calibration function, it is increased to
the minimum needed value.`)
	par.calPeaks = flag.Int("calmult",
		10,
		`only the topmost (<calmult> * <number of potential calibrants>)
peaks are considered for computing the recalibration. <1 means all peaks.`)
	par.minPeak = flag.Float64("minpeak",
		0.0,
		`minimum peak intensity to consider for computing the recalibration.`)
	par.rtWindow = flag.String("rt",
		"-10.0:10.0",
		`rt window (s)`)
	par.mzErrPPM = flag.Float64("ppmuncal",
		10.0,
		`max mz error (ppm) for trying to use calibrant for calibration`)
	par.mzTargetPPM = flag.Float64("ppmcal",
		2.0,
		`max mz error (ppm) for accepting a calibrant for calibration`)
	par.scoreFilter = flag.String("scorefilter",
		"MS:1002466(0.99:)MS:1002257(0.0:1e-2)MS:1001159(0.0:1e-2)",
		`filter for PSM scores to accept. Format:
<CVterm1|scorename1>([<minscore1>]:[<maxscore1>])...
When multiple score names/CV terms are specified, the first one on the list
that matches a score in the input file will be used.
TODO: The default contains reasonable values for some common search engines
and post-search scoring software:
  MS:1002257 (Comet:expectation value)
  MS:1001159 (SEQUEST:expectation value)
  MS:1002466 (PeptideShaker PSM score)`)
	par.charge = flag.String("charge",
		"1:5",
		`charge range of calibrants, or the string "ident". If set to "ident",
only the charge as found in the mzIdentMl file will be used for calibration.`)
	version := flag.Bool("version", false,
		`Show software version`)
	flag.Usage = usage
	flag.Parse()
	if *version {
		fmt.Fprintf(os.Stderr, "%s version %s\n", progName, progVersion)
		return
	}
	par.args = flag.Args()
	// Check if debug output should be enabled
	par.debug = os.Getenv("MZRECAL_DEBUG") == `1`

	sanatizeParams(&par)
	switch *par.stage {
	case 1:
		makeRecalCoefficients(par)
	case 2:
		doRecal(par)
	default:
		mzML, recal := makeRecalCoefficients(par)
		calibMzML(par, mzML, recal)
	}
}
