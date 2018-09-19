// This file contains code to help debugging, and is
// separated in from the rest in order not to litter
// the main code with debugging stuff

package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"mzrecal/mzml"
	"regexp"
	"strconv"
)

var debugSpecs *string // Print debug output for given spectrum range

// Parse string like "-12:6" into 2 values, -12 and 6
// Parameters min and max are the "default" min/max values,
// when a value is not specified (e.g. "-12:"), the defauls is assigned
func parseIntRange(r string, min int, max int) (int, int, error) {
	re := regexp.MustCompile(`\s*(\-?\d*):(\-?\d*)`)
	m := re.FindStringSubmatch(r)
	minOut := min
	maxOut := max
	if m[1] != "" {
		minOut, _ = strconv.Atoi(m[1])
		if minOut < min {
			minOut = min
		}
	}
	if m[2] != "" {
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

func init() {
	debugSpecs = flag.String("debug", "",
		`Print debug output for given spectrum range e.g. 3:6`)
}

func debugLogSpecs(i int, numSpecs int, retentionTime float64,
	peaks []mzml.Peak, mzMatchingCals []mzCalibrant, par params) {

	if *debugSpecs != `` {
		debugMin, debugMax, _ := parseIntRange(*debugSpecs, 0, numSpecs)
		if i >= debugMin && i <= debugMax {
			mz2matchIndex := make(map[float64]int)
			for k, mzMatch := range mzMatchingCals {
				mz2matchIndex[mzMatch.mzMeasured] = k
			}
			fmt.Printf("Spectrum:%d rt:%f\n", i, retentionTime)
			for j, p := range peaks {
				fmt.Printf("%d mzMeas:%f intens:%f", j, p.Mz, p.Intens)
				k, exists := mz2matchIndex[p.Mz]
				if exists {
					var rtShift, rtRel float64
					// Compounds that elute at all times have no valid retention time
					if mzMatchingCals[k].cal.retentionTime != -math.MaxFloat64 {
						rtShift = retentionTime - mzMatchingCals[k].cal.retentionTime
						if rtShift < 0 {
							rtRel = rtShift / *par.lowRT
						} else {
							rtRel = rtShift / *par.upRT
						}
					}
					mzRel := 100000000.0 * (1.0 - mzMatchingCals[k].mz/p.Mz) / (*par.mzErrPPM)
					fmt.Printf(" mzComp:%f(%0.2f%%) cal:%s rtShift:%f(%0.2f%%) charge:%d",
						mzMatchingCals[k].mz, mzRel,
						mzMatchingCals[k].cal.name,
						rtShift, rtRel*100,
						mzMatchingCals[k].charge)
				}
				fmt.Printf("\n")
			}
		}
	}
}
