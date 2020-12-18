// This file contains code to help debugging, and is
// separated in from the rest in order not to litter
// the main code with debugging stuff

package main

import (
	"flag"
	"fmt"
	"math"
	"sync"

	"github.com/524D/mzrecal/internal/mzml"
)

var debugSpecs *string // Print debug output for given spectrum range

// Map each calibrant to it's used spectrum id's
var calUsed4Spec = map[identifiedCalibrant][]int{}
var calUsed4SpecMux sync.Mutex

func init() {
	debugSpecs = flag.String("debug", "",
		"Print debug output for given spectrum `range` e.g. 3:6")
}

func debugLogSpecs(i int, numSpecs int, retentionTime float64,
	peaks []mzml.Peak, matchingCals []calibrant, par params,
	calibrantsUsed []int, recalMethod int, specRecalPar specRecalParams) {

	if *debugSpecs != `` {
		debugMin, debugMax, _ := parseIntRange(*debugSpecs, 0, numSpecs)
		if i >= debugMin && i <= debugMax {
			mz2matchIndex := make(map[float64]int)
			for k, mzMatch := range matchingCals {
				mz2matchIndex[mzMatch.mzMeasured] = k
			}
			isCalibrantUsed := make([]bool, len(matchingCals))
			for _, j := range calibrantsUsed {
				isCalibrantUsed[j] = true
			}

			recalPar := setRecalPars(recalMethod, specRecalPar)

			fmt.Printf("Spectrum:%d rt:%f\n", i, retentionTime)
			var mzRecalRelSum float64
			var mzRecalRelCount int
			for j, p := range peaks {
				fmt.Printf("%d mzMeas:%f intens:%f", j, p.Mz, p.Intens)
				k, exists := mz2matchIndex[p.Mz]

				if exists {
					mzMatchingCal := matchingCals[k]
					mzRel := 100000000.0 * (p.Mz/mzMatchingCal.mz - 1.0) / (*par.mzErrPPM)
					mzRecal := mzRecal(p.Mz, &recalPar)
					mzRecalRel := 100000000.0 * (mzRecal/mzMatchingCal.mz - 1.0) / (*par.mzTargetPPM)
					used := `-`
					if isCalibrantUsed[k] {
						used = `+`
						mzRecalRelSum += mzRecalRel
						mzRecalRelCount++
					}
					fmt.Printf(" mzCalc:%f(%0.2f%%) mzRecal:%f(%0.2f%%) used: %s [",
						matchingCals[k].mz, mzRel,
						mzRecal, mzRecalRel,
						used)

					for _, chargedCal := range mzMatchingCal.chargedCals {
						idCal := chargedCal.idCal

						var rtShift, rtRel float64
						// Compounds that elute at all times have no valid retention time
						// FIXME: We should consider retention times of all merged calibrants
						if idCal.retentionTime != -math.MaxFloat64 {
							rtShift = retentionTime - idCal.retentionTime
							if rtShift < 0 {
								rtRel = 100.0 * rtShift / -par.lowRT
							} else {
								rtRel = 100.0 * rtShift / par.upRT
							}
						}
						fmt.Printf(" cal:%s rtShift:%f(%0.2f%%) mz: %f charge:%d id-charge: %d;",
							idCal.name,
							rtShift,
							rtRel,
							chargedCal.mz,
							chargedCal.charge,
							idCal.idCharge)
					}
					fmt.Printf("]")
				}
				fmt.Printf("\n")
			}
			if mzRecalRelCount > 0 {
				fmt.Printf("Mean relative error after recalibration of accepted calibrants: %0.2f%% of target\n",
					mzRecalRelSum/float64(mzRecalRelCount))
			}
		}
	}
}

func debugLogPrecursorUpdate(i int, numSpecs int, mzOrig float64, mzNew float64, par params) {

	if *debugSpecs != `` {
		debugMin, debugMax, _ := parseIntRange(*debugSpecs, 0, numSpecs)
		if i >= debugMin && i <= debugMax {
			fmt.Printf("Spec %d precursor changed from %f to %f (%f)\n",
				i, mzOrig, mzNew, mzOrig-mzNew)
		}
	}
}

func debugRegisterCalUsed(i int, matchingCals []calibrant, par params, calibrantsUsed []int) {
	for _, cu := range calibrantsUsed {
		mc := matchingCals[cu]
		calUsed4SpecMux.Lock()
		for _, cc := range mc.chargedCals {
			calUsed4Spec[*cc.idCal] = append(calUsed4Spec[*cc.idCal], i)
		}
		calUsed4SpecMux.Unlock()
	}
}

func debugListUnusedCalibrants(allCals []identifiedCalibrant) {
	if *debugSpecs != `` {
		fmt.Printf("Unused calibrants\n")
		calUsed4SpecMux.Lock()
		for _, ac := range allCals {
			_, ok := calUsed4Spec[ac]
			if !ok {
				fmt.Printf("%+v mz:%f\n", ac, (ac.mass+float64(ac.idCharge)*massProton)/float64(ac.idCharge))
			}
		}
		calUsed4SpecMux.Unlock()
	}
}
