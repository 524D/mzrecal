package main

// // Dynamic link against gsl (probably needed for Cygwin)
// // #cgo LDFLAGS: -lgsl -lgslcblas -lm

// Static link against gls

/*
#cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgsl.a
#cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgslcblas.a
#cgo LDFLAGS: -lm
#include <stdlib.h>
#include <stdio.h>
#include <gsl/gsl_vector.h>
#include <gsl/gsl_blas.h>
#include <gsl/gsl_multifit_nlin.h>

// Based on MSRECAL/RECAL_FUNCTIONS.C

// WARNING: the following enum must stay consistent with the iota in mzrecal.go
typedef enum {
    CALIB_NONE,
    CALIB_FTICR,
    CALIB_TOF,
    CALIB_ORBITRAP,
} calib_method_t;

// FIXME: Change this constant to a parameter
#define INTERNAL_CALIBRATION_TARGET 3e-6	//discard internal calibrants that do not fit CAL2 better than this

// The maximum number of calibration parameters of any calibration function
#define MAX_CAL_PARS 10

// Maximum number of iterations fro the FDF solver
#define MAX_FDF_SOLVER_ITER 500

// The cal_params_t contains the result of the recalibration
typedef struct {
    calib_method_t calib_method;  // The calibration method
    int            nr_cal_pars;   // Number of parameters for this method
    double         cal_pars[MAX_CAL_PARS]; // The actual parameters
    int            n_calibrants;  // Number of calibrants used for result
} cal_params_t;

// Calibrant description type
typedef struct {
    double mz_calc; // calculated m/z
    double mz_meas; // measured m/z
} calibrant_t;

typedef struct {
    calib_method_t calib_method;
    int            n_calibrants;
    calibrant_t *  calibrants;
} recal_data_t;

double mz_recalX(double mz_meas, cal_params_t *p)
{
    double mz_calib;
    switch (p->calib_method) {
    case CALIB_FTICR:
        // mz_calib = Ca/((1/mz_meas)-Cb)
        mz_calib = (p->cal_pars[0])/((1/mz_meas)-(p->cal_pars[1]));
        break;
    case CALIB_TOF: // FIXME: implement correct calib function
        mz_calib = (p->cal_pars[0])/((1/mz_meas)-(p->cal_pars[1]));
        break;
    case CALIB_ORBITRAP: { // FIXME: implement correct calib function
        // mz_calib = A/(f*f) = A / (1/sqrt(mz_meas) * 1/sqrt(mz_meas))
        //          = A*mz_meas
        mz_calib = p->cal_pars[0]*mz_meas;
        break;
    }
    default:
        mz_calib = mz_meas;
        break;
    }
    return mz_calib;
}

int calib_f(const gsl_vector *x, void *params, gsl_vector *f)
{
    recal_data_t *recal_data_p = (recal_data_t *)params;

    calibrant_t *calibrants = recal_data_p->calibrants;
    int n_calibrants = recal_data_p->n_calibrants;

    switch (recal_data_p->calib_method) {
    case CALIB_FTICR: {
        double a = gsl_vector_get (x, 0);
        double b = gsl_vector_get (x, 1);
        double M;
        double freq;
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            // Model m = a/(f-b) (CAL2 inverted)
            freq= 1/calibrants[i].mz_meas;
            M = a/(freq-b);
            gsl_vector_set (f, i, (M-calibrants[i].mz_calc)); // absolute or relative error?
        }
        break;
    }
    case CALIB_TOF: { // FIXME: implement correct function
        double a = gsl_vector_get (x, 0);
        double b = gsl_vector_get (x, 1);
        double M;
        double freq;
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            // Model m = a/(f-b) (CAL2 inverted)
            freq= 1/calibrants[i].mz_meas;
            M = a/(freq-b);
            gsl_vector_set (f, i, (M-calibrants[i].mz_calc)); // absolute or relative error?
        }
        break;
    }
    case CALIB_ORBITRAP: { // FIXME: implement correct function
        double a = gsl_vector_get (x, 0);
        double mz_calib;
        double freq;
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            // Model m = a/(f*f) (CAL2 inverted)
            //         = a*mz_meas
            mz_calib = a*calibrants[i].mz_meas;
            gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
        }
        break;
     }
     default:
         break;
     }
     return GSL_SUCCESS;
}

// DF calibrator
int calib_df(const gsl_vector *x, void *params, gsl_matrix *J)
{
    recal_data_t *recal_data_p = (recal_data_t *)params;

    calibrant_t *calibrants = recal_data_p->calibrants;
    int n_calibrants = recal_data_p->n_calibrants;

    switch (recal_data_p->calib_method) {
    case CALIB_FTICR: {
        double a = gsl_vector_get (x, 0);
        double b = gsl_vector_get (x, 1);
        double freq;
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            freq=1/calibrants[i].mz_meas;
            gsl_matrix_set (J,i,0, 1/(freq-b) );
            gsl_matrix_set (J,i,1, a/((freq-b)*(freq-b)) );
        }
        break;
    }
    case CALIB_TOF: { // FIXME: implement correct function
        double a = gsl_vector_get (x, 0);
        double b = gsl_vector_get (x, 1);
        double freq;
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            freq=1/calibrants[i].mz_meas;
            gsl_matrix_set (J,i,0, 1/(freq-b) );
            gsl_matrix_set (J,i,1, a/((freq-b)*(freq-b)) );
        }
       break;
    }
    case CALIB_ORBITRAP: { // FIXME: implement correct function
        double a = gsl_vector_get (x, 0);
        size_t i;

        for (i=0;i<n_calibrants;i++) {
            gsl_matrix_set (J,i,0, calibrants[i].mz_meas );
        }
        break;
    }
    default:
        break;
    }
    return GSL_SUCCESS;
}

// FDF Calibrator
int calib_fdf(const gsl_vector *x, void *params, gsl_vector *f, gsl_matrix *J)
{
     calib_f (x,params,f);
     calib_df (x,params,J);
     return GSL_SUCCESS;
}

void init_cal_params(cal_params_t *cal_params, calib_method_t calib_method) {
    cal_params->calib_method = calib_method;
    // Initialize calibration paramters close to the optimum
    switch (calib_method) {
    case CALIB_FTICR:
        cal_params->cal_pars[0] = 1.0;
        cal_params->cal_pars[1] = 0.0;
        cal_params->nr_cal_pars = 2;
        break;
    case CALIB_TOF:
    // FIXME: fill in correct init parameters
        cal_params->cal_pars[0] = 1.0;
        cal_params->cal_pars[1] = 0.0;
        cal_params->nr_cal_pars = 2;
        break;
    case CALIB_ORBITRAP:
        cal_params->cal_pars[0] = 1.0;
        cal_params->nr_cal_pars = 1;
        break;
    default:
     		cal_params->nr_cal_pars = 0;
    break;
    }
}

cal_params_t recalibratePeaks(recal_data_t d,
                                   int min_cal, int spec_nr){
    int status, satisfied, j, vi;

    const gsl_multifit_fdfsolver_type *T;
    gsl_multifit_fdfsolver *s;
    double chi;
    int iter=0;
    cal_params_t cal_params;

    init_cal_params(&cal_params, d.calib_method);

    gsl_multifit_function_fdf func;

    // ??? The array fed to gsl_vector_view_array needs to be a copy
    // otherwise the result is not the same
    cal_params_t cal_params_copy = cal_params;
    gsl_vector_view x=gsl_vector_view_array(cal_params_copy.cal_pars,cal_params.nr_cal_pars);

    func.f = &calib_f;
    func.df = &calib_df;
    func.fdf = &calib_fdf;

    satisfied=0;
    while (d.n_calibrants >= min_cal && !satisfied) {
        // least-squares fit first using all peaks, than removing those that don't fit
        iter=0;
        T = gsl_multifit_fdfsolver_lmder;
        s = gsl_multifit_fdfsolver_alloc (T, d.n_calibrants, cal_params.nr_cal_pars);
        func.n = d.n_calibrants;
        func.p = cal_params.nr_cal_pars;
        func.params = &d;
        gsl_multifit_fdfsolver_set(s,&func,&x.vector);

        do {
            iter++;
            status = gsl_multifit_fdfsolver_iterate (s);

            if (status)
                break;
            status=gsl_multifit_test_delta (s->dx, s->x, 1e-9, 1e-9);
        } while (status==GSL_CONTINUE && iter<MAX_FDF_SOLVER_ITER);

        for (vi=0; vi<cal_params.nr_cal_pars; vi++) {
            cal_params.cal_pars[vi] = gsl_vector_get(s->x,vi);
        }

        chi = gsl_blas_dnrm2(s->f);
        gsl_multifit_fdfsolver_free(s);

        // OK, that was one internal recalibration, now lets check if all calibrants are < INTERNAL_CALIBRATION_TARGET, if not, throw these out
        // and recalibrate (as long as we have at least min_cal peaks)
        int accepted_idx = 0;
        for(j=0; j<d.n_calibrants; j++) {
            double mz_calc = d.calibrants[j].mz_calc;
            double mz_meas = d.calibrants[j].mz_meas;
            double mz_recal = mz_recalX(mz_meas, &cal_params);
            if (fabs((mz_calc-mz_recal)/mz_calc)<INTERNAL_CALIBRATION_TARGET) {
                d.calibrants[accepted_idx++] = d.calibrants[j];
            }
        }
        // If all (remaining) calibrants are accepted, we are done
        if (accepted_idx == d.n_calibrants) {
            satisfied=1; // all calibrants < INTERNAL_CALIBRATION_TARGET
        }
        d.n_calibrants=accepted_idx;
    }
    cal_params.n_calibrants = d.n_calibrants;
    if (!satisfied) {
        cal_params.nr_cal_pars = -1;
        cal_params.n_calibrants = 0;
    }
    return cal_params;
}

// Function fill_calibrant_list is only needed because directly filling
// a C "pointer array" from Go is a bit messy.
void fill_calibrant_list(calibrant_t *calibrant_list, int i,
                         double mz_calc, double mz_measured) {
    calibrant_list[i].mz_calc = mz_calc;
    calibrant_list[i].mz_meas = mz_measured;
}
*/
import "C"
import (
	"unsafe"
)

func recalibrateSpec(specNr int, recalMethod int,
	mzCalibrants []mzCalibrant, par params) (
	specRecalParams, int, error) {
	var specRecalPar specRecalParams

	specRecalPar.SpecNr = specNr

	// FIXME: Handle out of memory for C.malloc (not sure if it returns nil or panics...)
	calibrantList := (*C.calibrant_t)(C.malloc(C.ulong(C.sizeof_calibrant_t * len(mzCalibrants))))
	for i, calibrant := range mzCalibrants {
		C.fill_calibrant_list(calibrantList, C.int(i), C.double(calibrant.mz), C.double(calibrant.mzMeasured))
	}
	var recalData C.recal_data_t
	recalData.calib_method = C.calib_method_t(recalMethod)
	recalData.n_calibrants = C.int(len(mzCalibrants))
	recalData.calibrants = calibrantList
	specCalResult, _ := C.recalibratePeaks(recalData,
		C.int(*par.minCal), C.int(specNr))
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
