package main

// // Dynamic link against gsl (probably needed for Cygwin)
// // #cgo LDFLAGS: -lgsl -lgslcblas -lm

// Static link against gls

/*
#cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgsl.a
#cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgslcblas.a
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include "recal.h"
*/
import "C"
import (
	"unsafe"
)

func recalibrateSpec(specIndex int, recalMethod int,
	mzCalibrants []mzCalibrant, par params) (
	specRecalParams, int, error) {
	var specRecalPar specRecalParams

	specRecalPar.SpecIndex = specIndex

	// FIXME: Handle out of memory for C.malloc (not sure if it returns nil or panics...)
	calibrantList := (*C.calibrant_t)(C.malloc(C.ulong(C.sizeof_calibrant_t * len(mzCalibrants))))
	for i, calibrant := range mzCalibrants {
		C.fill_calibrant_list(calibrantList, C.int(i), C.double(calibrant.mz), C.double(calibrant.mzMeasured))
	}
	var recalData C.recal_data_t
	recalData.calib_method = C.calib_method_t(recalMethod)
	recalData.n_calibrants = C.int(len(mzCalibrants))
	recalData.calibrants = calibrantList
	// Aim to calibrate to half the specified mass error
	internalCalibrationTarget := *par.mzTargetPPM / 1000000.0
	specCalResult, _ := C.recalibratePeaks(recalData, C.int(*par.minCal),
		C.double(internalCalibrationTarget), C.int(specIndex))
	C.free(unsafe.Pointer(calibrantList))
	for i := 0; i < int(specCalResult.nr_cal_pars); i++ {
		specRecalPar.P = append(specRecalPar.P, float64(specCalResult.cal_pars[i]))
	}
	calibrantsUsed := int(specCalResult.n_calibrants)
	return specRecalPar, calibrantsUsed, nil
}

func setRecalPars(recalMethod int, specRecalPar specRecalParams) C.cal_params_t {
	var cCalPar C.cal_params_t
	cCalPar.calib_method = C.calib_method_t(recalMethod)
	cCalPar.nr_cal_pars = C.int(len(specRecalPar.P))
	for i, p := range specRecalPar.P {
		cCalPar.cal_pars[i] = C.double(p)
	}
	return cCalPar
}

func mzRecal(mz float64, recalPar *C.cal_params_t) float64 {
	mzNew := float64(C.mz_recalX(C.double(mz), recalPar))
	return mzNew
}
