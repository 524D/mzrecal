package main

/*
#include <stdlib.h>
#include "recal.h"
*/
import "C"
import (
	"unsafe"
)

// CalParams packs the "c" structure with calibration parameters
type CalParams struct {
	cCalPar C.cal_params_t
}

func recalibrateSpec(specIndex int, recalMethod int,
	mzCalibrants []calibrant, par params) (
	specRecalParams, []int, error) {
	var specRecalPar specRecalParams

	specRecalPar.SpecIndex = specIndex

	// FIXME: Handle out of memory for C.malloc (not sure if it returns nil or panics...)
	calibrantList := (*C.calibrant_t)(C.malloc(C.size_t(C.sizeof_calibrant_t * len(mzCalibrants))))
	for i, calibrant := range mzCalibrants {
		C.fill_calibrant_list(calibrantList, C.int(i), C.double(calibrant.mz),
			C.double(calibrant.mzMeasured))
	}
	var recalData C.recal_data_t
	recalData.calib_method = C.calib_method_t(recalMethod)
	recalData.n_calibrants = C.int(len(mzCalibrants))
	recalData.calibrants = calibrantList
	// Aim to calibrate to half the specified mass error
	internalCalibrationTarget := *par.mzTargetPPM / 1000000.0
	cDebug := 0
	//	if specIndex >= 1994 && specIndex <= 2000 {
	//	cDebug = 1
	//	}
	specCalResult, _ := C.recalibratePeaks(&recalData, C.int(*par.minCal),
		C.double(internalCalibrationTarget), C.int(specIndex), C.int(cDebug))
	// make a slice of calibrant indexes which where used for calibration
	calibrantsUsed := make([]int, 0, int(specCalResult.n_calibrants))
	for i := 0; i < int(specCalResult.n_calibrants); i++ {
		calibrantsUsed = append(calibrantsUsed,
			int(C.get_calibrant_id(calibrantList, C.int(i))))
	}
	C.free(unsafe.Pointer(calibrantList))
	for i := 0; i < int(specCalResult.nr_cal_pars); i++ {
		specRecalPar.P = append(specRecalPar.P, float64(specCalResult.cal_pars[i]))
	}
	return specRecalPar, calibrantsUsed, nil
}

func setRecalPars(recalMethod int, specRecalPar specRecalParams) CalParams {
	var calPar CalParams
	calPar.cCalPar.calib_method = C.calib_method_t(recalMethod)
	calPar.cCalPar.nr_cal_pars = C.int(len(specRecalPar.P))
	for i, p := range specRecalPar.P {
		calPar.cCalPar.cal_pars[i] = C.double(p)
	}
	return calPar
}

func mzRecal(mz float64, recalPar *CalParams) float64 {
	mzNew := float64(C.mz_recalX(C.double(mz), &recalPar.cCalPar))
	return mzNew
}

func getNrCalPars(recalMethod int) int {
	return int(C.get_nr_cal_pars(C.calib_method_t(recalMethod)))
}
