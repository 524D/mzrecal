// Copyright 2017 Rob Marissen. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"math"
	"os"
	"sort"

	"msrecal/mzidentml"
	"msrecal/mzml"
)

const mergeMassTol = float64(1e-7)
const protonMass = float64(1.007276466879)

type params struct {
	mzMLFilename      *string
	mzIdentMlFilename *string
	calFilename       *string  // Filename where JSON calibration parameters will be written
	minCal            *int     // minimum number of calibrants a spectrum should have to be recalibrated
	minPeak           *float64 // minimum intensity of peaks to be considered for recalibrating
	lowRT             *float64 // lower rt window boundary
	upRT              *float64 // upper rt window boundary
	mzErrPPM          *float64 // max mass measurement error
	scoreName         *string  // the name of the score parameter
	minScore          *float64 // min score filter
	maxScore          *float64 // max score filter
	minMz             *int     // min m/z for calibrants
	maxMz             *int     // max m/z for calibrants
}

type calibrant struct {
	name          string
	mass          float64 // Uncharged mass
	retentionTime float64
	singleCharged bool
}

type specRecalParams struct {
	SpecNr int
	P      []float64
}

// recalParams contains recalibration parameters for each spectrum,
// in addition to generic recalibration data for the whole file
type recalParams struct {
	// Version of recalibration parameters, used when storing/loading
	// parameters in JSON format for different version of the software
	MSRecalVersion string
	SpecRecalPar   []specRecalParams
}

type mzCalibrant struct {
	mz         float64    // computed mz of the calibrant
	mzMeasured float64    // mz of the best candidate peak
	cal        *calibrant // Only used for verbose output
	charge     int        // Only used for verbose output
}

var fixedCalibrants = []calibrant{

	// cyclosiloxanes, H6nC2nOnSin
	calibrant{
		name:          `cyclosiloxane8`,
		mass:          592.1503308,
		retentionTime: -math.MaxFloat64, // Indicates any retention time
		singleCharged: true,
	},
	calibrant{
		name:          `cyclosiloxane9`,
		mass:          666.1691221,
		retentionTime: -math.MaxFloat64,
		singleCharged: true,
	},
	calibrant{
		name:          `cyclosiloxane10`,
		mass:          740.1879134,
		retentionTime: -math.MaxFloat64,
		singleCharged: true,
	},
	calibrant{
		name:          `cyclosiloxane11`,
		mass:          814.2067048,
		retentionTime: -math.MaxFloat64,
		singleCharged: true,
	},
	calibrant{
		name:          `cyclosiloxane12`,
		mass:          888.2254961,
		retentionTime: -math.MaxFloat64,
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

// The following are needed for sorting []calibrant on mass and retentionTime
// the mass field.
type byMass []calibrant

func (a byMass) Len() int           { return len(a) }
func (a byMass) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byMass) Less(i, j int) bool { return a[i].mass < a[j].mass }

type byRetention []calibrant

func (a byRetention) Len() int           { return len(a) }
func (a byRetention) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byRetention) Less(i, j int) bool { return a[i].retentionTime < a[j].retentionTime }

type peaksByMass []mzml.Peak

func (a peaksByMass) Len() int           { return len(a) }
func (a peaksByMass) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a peaksByMass) Less(i, j int) bool { return a[i].Mz < a[j].Mz }

// Compute the smallest isotopic mass of a peptide
func pepMass(pepSeq string) (float64, error) {
	m := float64(18.0105647) // H2O
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
// For each calibrant, it:
// - computes the mass of the lightest isotope
// - get the retention name, retentionTime, spectrum
func makeCalibrantList(mzIdentML *mzidentml.MzIdentML, par params) (
	[]calibrant, error) {
	// Create slice for the numbar of calibrants that we expect to have
	cals := make([]calibrant, 0, mzIdentML.NumIdents()+len(fixedCalibrants))
	for i := 0; i < mzIdentML.NumIdents(); i++ {
		ident, err := mzIdentML.Ident(i)
		if err != nil {
			return nil, err
		}
		var cal calibrant
		m, err := pepMass(ident.PepSeq)
		if err == nil { // Skip if mass cannot be computed
			cal.name = ident.PepID
			cal.retentionTime = ident.RetentionTime
			cal.singleCharged = false
			cal.mass = m + ident.ModMass
			cals = append(cals, cal)
		}
	}
	cals = append(cals, fixedCalibrants...)
	sort.Sort(byRetention(cals))

	return cals, nil
}

func calibsInRtWindows(rtMin, rtMax float64, allCals []calibrant) ([]calibrant, error) {

	// Find the indexes of the calibrants within the retention time window
	i1 := sort.Search(len(allCals), func(i int) bool { return allCals[i].retentionTime >= rtMin })
	i2 := sort.Search(len(allCals), func(i int) bool { return allCals[i].retentionTime > rtMax })

	// Find calibrants that elute at all retention times
	// These have elusiton time -math.MaxFloat64
	var i3 int
	for i3 = 0; allCals[i3].retentionTime == -math.MaxFloat64; i3++ {
	}

	var cals = make([]calibrant, 0, (i2-i1)+i3)
	cals = append(cals, allCals[i1:i2]...)
	cals = append(cals, allCals[0:i3]...)

	return cals, nil
}

// mergeSameMassCals merges all calibrants that have the same mass or
// nearly the same mass. The mass of the first calibrant is retained
func mergeSameMassCals(cals []calibrant) []calibrant {
	mcals := make([]calibrant, 0, len(cals))
	// sort calibrants by mass
	sort.Sort(byMass(cals))

	prevMass := float64(-1)
	for _, cal := range cals {
		if math.Abs(cal.mass-prevMass) < mergeMassTol {
			mcals[len(mcals)-1].name += `;` + cal.name
		} else {
			mcals = append(mcals, cal)
		}
		prevMass = cal.mass
	}
	return mcals
}

func newMzCalibrant(charge int, cal *calibrant) mzCalibrant {
	var mzCal mzCalibrant

	fCharge := float64(charge)
	mzCal.mz = (cal.mass + fCharge*protonMass) / fCharge
	mzCal.cal = cal
	mzCal.charge = charge
	return mzCal
}

func recalibrate(mzML *mzml.MzML, cals []calibrant, par params) (recalParams, error) {
	var recal recalParams

	recal.MSRecalVersion = "1.0"
	for i := 0; i < mzML.NumSpecs(); i++ {
		msLevel, err := mzML.MSLevel(i)
		if err != nil {
			return recal, err
		}
		if msLevel == 1 {
			retentionTime, err := mzML.RetentionTime(i)
			if err != nil {
				return recal, err
			}
			wCals, err := calibsInRtWindows(retentionTime-*par.lowRT,
				retentionTime+*par.upRT, cals)
			if err != nil {
				return recal, err
			}
			specCals := mergeSameMassCals(wCals)

			// Make slice with mz values for all calibrants
			// For efficiency, pre-allocate (more than) enough elements
			mzCalibrants := make([]mzCalibrant, 0,
				len(specCals)*(*par.maxMz-*par.minMz))
			for j, cal := range specCals {
				//				log.Printf("Calibrating spec %d, rt %f, calibrants: %+v\n", j, retentionTime, cal)
				if cal.singleCharged {
					mzCalibrants = append(mzCalibrants, newMzCalibrant(1, &specCals[j]))
				} else {
					for charge := *par.minMz; charge <= *par.maxMz; charge++ {
						mzCalibrants = append(mzCalibrants, newMzCalibrant(charge, &specCals[j]))
					}
				}
			}

			peaks, err := mzML.ReadScan(i)
			if err != nil {
				log.Fatalf("recalibrateSpec ReadScan failed for spectrum %d: %v",
					i, err)
			}
			mzMatchingCals := mzCalibrantsMatchPeaks(peaks, mzCalibrants, par)
			log.Printf("%d nr mzCalibrants: %d mzMatchingCals %d",
				i, len(mzCalibrants), len(mzMatchingCals))

			specRecalPar, err := recalibrateSpec(i, mzMatchingCals)
			if err != nil {
				log.Printf("recalibrateSpec calibration failed for spectrum %d: %v",
					i, err)
			} else {

			}
			recal.SpecRecalPar = append(recal.SpecRecalPar, specRecalPar)
		}
	}

	return recal, nil
}

func mzCalibrantsMatchPeaks(peaks []mzml.Peak, mzCalibrants []mzCalibrant, par params) []mzCalibrant {
	mzMatchingCals := make([]mzCalibrant, 0, len(mzCalibrants))

	// For each potental calibrant, find highest peak within mz window

	// Peaks in mzml probably always are sorted by mass, but that is not specified
	// by the schema/mzML description. Therefore, we must sort them now.
	sort.Sort(peaksByMass(peaks))

	for _, mzCalibrant := range mzCalibrants {
		mz := mzCalibrant.mz
		mzErr := *par.mzErrPPM * mz / 1000000.0
		peak := maxPeakInMzWindow(mz-mzErr, mz+mzErr, peaks)
		if peak.Intens > *par.minPeak {
			mzCalibrant.mzMeasured = peak.Mz
			mzMatchingCals = append(mzMatchingCals, mzCalibrant)
		}
	}
	return mzMatchingCals
}

// maxPeakInMzWindow returns the highest initensity peak in a given mz window.
// Peaks must be ordered by mz prior to calling this function
// If no peak was found, peak.intensity will be 0
func maxPeakInMzWindow(mzMin, mzMax float64, peaks []mzml.Peak) mzml.Peak {

	// Find the indexes of the calibrants within the retention time window
	i1 := sort.Search(len(peaks), func(i int) bool { return peaks[i].Mz >= mzMin })
	i2 := sort.Search(len(peaks), func(i int) bool { return peaks[i].Mz > mzMax })

	// if i1 < len(peaks) && i2 < len(peaks) {
	// 	log.Printf("cnt %d peaks[i1].Mz %f, mzMin %f, peaks[i2].Mz %f, mzMax %f\n", i2-i1, peaks[i1].Mz, mzMin, peaks[i2].Mz, mzMax)
	// }

	var peak mzml.Peak // auto initialzed to 0.0, 0.0
	for i := i1; i < i2; i++ {
		if peaks[i].Intens > peak.Intens {
			peak = peaks[i]
		}
	}
	return peak
}

func recalibrateSpec(specNr int, mzCalibrants []mzCalibrant) (specRecalParams, error) {
	var specRecalPar specRecalParams

	specRecalPar.SpecNr = specNr

	return specRecalPar, nil
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

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var par params

	par.mzMLFilename = flag.String("mzml",
		"test.mzml",
		"mzML filename")
	par.mzIdentMlFilename = flag.String("mzid",
		"test.mzid",
		"mzIdentMl filename")
	par.calFilename = flag.String("cal",
		"test.cal.json",
		"filename for calibration paramters")
	par.minCal = flag.Int("mincals",
		3,
		"minimum number of calibrants a spectrum should have to be recalibrated")
	par.minPeak = flag.Float64("minPeak",
		10000,
		"minimum intensity of peaks to be considered for recalibrating")
	par.lowRT = flag.Float64("lowRT",
		10,
		"lower rt window boundary (s)")
	par.upRT = flag.Float64("upRT",
		10,
		"upper rt window boundary (s)")
	par.mzErrPPM = flag.Float64("massErr",
		10.0,
		"max mz error for assigning a peak to a calibrant")
	par.scoreName = flag.String("scoreName",
		"MS-GF:PepQValue",
		"the name of the score parameter")
	par.minScore = flag.Float64("minScore",
		0.1,
		"min score filter")
	par.maxScore = flag.Float64("maxScore",
		1.0,
		"min score filter")
	par.minMz = flag.Int("minmz",
		1,
		"min m/z for calibrants")
	par.maxMz = flag.Int("maxmz",
		5,
		"max m/z for calibrants")

	flag.Parse()
	f1, err := os.Open(*par.mzIdentMlFilename)
	if err != nil {
		log.Fatalf("Open: mzIdentMLfile %v", err)
	}
	defer f1.Close()
	mzIdentML, err := mzidentml.Read(f1)
	if err != nil {
		log.Fatalf("mzidentml.Read: error return %v", err)
	}
	cals, err := makeCalibrantList(&mzIdentML, par)
	if err != nil {
		log.Fatal("makeCalibrantList failed")
	}

	//	log.Printf("Calibrants (%d): %+v\n", len(cals), cals)
	f2, err := os.Open(*par.mzMLFilename)
	if err != nil {
		log.Fatalf("Open: mzMLfile %v", err)
	}
	defer f2.Close()
	mzML, err := mzml.Read(f2)
	if err != nil {
		log.Fatalf("mzml.Read: error return %v", err)
	}

	recal, err := recalibrate(&mzML, cals, par)
	if err != nil {
		log.Fatalf("recalibrate: error return %v", err)
	}

	err = writeRecal(recal, par)
	if err != nil {
		log.Fatalf("writeRecal: error return %v", err)
	}

	// FIXME: testcode below: remove
	gsl()
}
