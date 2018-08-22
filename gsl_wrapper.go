package main

// // Dynamic link against gsl (probably needed for Cygwin)
// // #cgo LDFLAGS: -lgsl -lgslcblas -lm

// Static link against gls

// #cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgsl.a
// #cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgslcblas.a
// #cgo LDFLAGS: -lm
// #include <stdlib.h>
// #include <stdio.h>
// #include <gsl/gsl_vector.h>
// #include <gsl/gsl_blas.h>
// #include <gsl/gsl_multifit_nlin.h>
//
// // COPIED/MODIFIED FROM MSRECAL/RECAL_FUNCTIONS.C
//
// typedef enum {
//    RECAL_FTICR,
//    RECAL_TOF,
//    RECAL_ORBITRAP
// } recal_method_t;
//
// // The maximum number of calibration parameters of any calibration function
// #define MAX_CAL_PARS 10
//
// typedef struct {
//     int satisfied;
//     int nr_cal_pars;
//     double cal_pars[MAX_CAL_PARS];
// } spec_cal_result_t;
//
// // Calibrant description type
// typedef struct {
//     double mz_calc; /* calculated m/z */
//     double mz_meas; /* measured m/z */
//     double intensity;
// } calibrant;
//
//
// typedef struct {
//     int      n_calibrants;
//     double * y; /* measured f (in phony units) */
//     double * mz2; /* theoretical m/z */
// } mz_data;
//
// // FIXME: Fixed size arrays: This sucks!
// #define MAX_CALIBRANTS 10000
// #define INTERNAL_CALIBRATION_TARGET 3e-6	//discard internal calibrants that do not fit CAL2 better than this
//
// double mz_recal(double mz_meas, double Ca, double Cb)
// {
// 	return Ca/((1/mz_meas)-Cb);
// }
//
// int calib_f(const gsl_vector *x, void *params, gsl_vector *f)
// {
// 	double *y = ((mz_data *)params)->y;
// 	double *mz = ((mz_data *)params)->mz2;
//  int n_calibrants = ((mz_data *)params)->n_calibrants;
// 	double a = gsl_vector_get (x, 0);
// 	double b = gsl_vector_get (x, 1);
// 	double M;
//     size_t i;
//
// 	for (i=0;i<n_calibrants;i++) {
// 		/* Model m = a/(f-b) (CAL2 inverted) */
// 		M = a/(y[i]-b);
// 		gsl_vector_set (f, i, (M-mz[i])); /* absolute or relative error? */
// 	}// for
//
//     return GSL_SUCCESS;
//
// }// int calib_f(const gsl_vector *x, void *params, gsl_vector *f)
//
// // DF calibrator
// int calib_df(const gsl_vector *x, void *params, gsl_matrix *J)
// {
// 	double *y = ((mz_data *)params)->y;
//  int n_calibrants = ((mz_data *)params)->n_calibrants;
// 	double a = gsl_vector_get (x, 0);
// 	double b = gsl_vector_get (x, 1);
// 	size_t i;
//
// 	for (i=0;i<n_calibrants;i++) {
// 		gsl_matrix_set (J,i,0, 1/(y[i]-b) );
// 		gsl_matrix_set (J,i,1, a/((y[i]-b)*(y[i]-b)) );
// 	}// for
//
// 	return GSL_SUCCESS;
//
// }// int calib_df (const gsl_vector *x, void *params, gsl_matrix *J)
//
// // FDF Calibrator
// int calib_fdf(const gsl_vector *x, void *params, gsl_vector *f, gsl_matrix *J)
// {
// 	calib_f (x,params,f);
// 	calib_df (x,params,J);
//
// 	return GSL_SUCCESS;
// }
//
// spec_cal_result_t recalibratePeaks(recal_method_t recal_method,
//                                    calibrant *calibrant_list,
//                                    int n_calibrants,
//                                    int min_cal, int spec_nr){
//     int status, SATISFIED, j;
//
//     const gsl_multifit_fdfsolver_type *T;
// 	   gsl_multifit_fdfsolver *s;
// 	   double chi;
//     int iter=0;
// 	   const size_t pp=2; /* number of free parameters in calibration function */
// 	   double y[MAX_CALIBRANTS];
// 	   double mz2[MAX_CALIBRANTS];
// 	   mz_data d={0,y,mz2};
// 	   double x_init[2]={1.0,0.0}; /* start here, close to minimum if reasonably calibrated beforehand */
//     double Ca, Cb;
//     spec_cal_result_t spec_cal_result;
//
//     gsl_multifit_function_fdf func;
//     gsl_vector_view x=gsl_vector_view_array(x_init,pp);
//     func.f = &calib_f;
//     func.df = &calib_df;
//     func.fdf = &calib_fdf;
//
//     SATISFIED=0;
//     while (n_calibrants >= min_cal && !SATISFIED) {
//     	   /* least-squares fit first using all peaks, than removing those that don't fit */
//         for (j=0;j<n_calibrants;j++) {
//             d.y[j] = 1 / calibrant_list[j].mz_meas;
//             d.mz2[j] = calibrant_list[j].mz_calc;
//         }
//         d.n_calibrants = n_calibrants;
//
//         iter=0;
//         T = gsl_multifit_fdfsolver_lmder;
//         s = gsl_multifit_fdfsolver_alloc (T, n_calibrants, pp); /* pp = 2 parameters, Ca and Cb */
//         func.n = n_calibrants;
//         func.p = pp;
//         func.params = &d;
//         gsl_multifit_fdfsolver_set(s,&func,&x.vector);
//
//         do {
//             iter++;
//             status = gsl_multifit_fdfsolver_iterate (s);
//
//             if (status)
//                 break;
//             status=gsl_multifit_test_delta (s->dx, s->x, 1e-9, 1e-9);
//         } while (status==GSL_CONTINUE && iter<500);
//
//         Ca = gsl_vector_get(s->x,0);
//         Cb = gsl_vector_get(s->x,1);
//         chi = gsl_blas_dnrm2(s->f);
//         gsl_multifit_fdfsolver_free(s);
//
//         /* OK, that was one internal recalibration, now lets check if all calibrants are < INTERNAL_CALIBRATION_TARGET, if not, throw these out */
//         /* and recalibrate (as long as we have at least min_cal peaks) */
//         int accepted_idx = 0;
//         for(j=0; j<n_calibrants; j++) {
//             if (fabs((calibrant_list[j].mz_calc-mz_recal(calibrant_list[j].mz_meas, Ca, Cb))/calibrant_list[j].mz_calc)<INTERNAL_CALIBRATION_TARGET) {
//                 calibrant_list[accepted_idx++] = calibrant_list[j];
//             }
//         }
//         if (accepted_idx == n_calibrants) {
//             SATISFIED=1; /* all calibrants < INTERNAL_CALIBRATION_TARGET */
//         }
//         n_calibrants=accepted_idx;
//     }
//     spec_cal_result.satisfied = SATISFIED;
//     if (SATISFIED) {
//         spec_cal_result.cal_pars[0] = Ca;
//         spec_cal_result.cal_pars[1] = Cb;
//         spec_cal_result.nr_cal_pars = 2;
//     }
//     else {
//         spec_cal_result.nr_cal_pars = 0;
//     }
// 	   return spec_cal_result;
// }
//
// // Function fill_calibrant_list is only needed because directly filling
// // a C "pointer array" from Go is a bit messy.
// void fill_calibrant_list(calibrant *calibrant_list, int i,
// 	 double mz_calc, double mz_measured) {
// 		 calibrant_list[i].mz_calc = mz_calc;
// 		 calibrant_list[i].mz_meas = mz_measured;
// }
import "C"
import (
	"errors"
	"unsafe"
)

func recalibrateSpec(specNr int, recalMethod string,
	mzCalibrants []mzCalibrant, par params) (specRecalParams, error) {
	var specRecalPar specRecalParams
	var recalMeth C.recal_method_t

	specRecalPar.SpecNr = specNr

	switch recalMethod {
	case `FTICR`:
		{
			recalMeth = C.RECAL_FTICR
		}
	case `TOF`:
		{
			recalMeth = C.RECAL_TOF
		}
	case `Orbitrap`:
		{
			recalMeth = C.RECAL_ORBITRAP
		}
	default:
		{
			return specRecalPar, errors.New("recalMethod invalid: " + recalMethod)
		}
	}
	// FIXME: Handle out of memory for malloc (not sure if it returns nil or panics...)
	calibrantList := (*C.calibrant)(C.malloc(C.ulong(C.sizeof_calibrant * len(mzCalibrants))))
	for i, calibrant := range mzCalibrants {
		C.fill_calibrant_list(calibrantList, C.int(i), C.double(calibrant.mz), C.double(calibrant.mzMeasured))
	}
	specCalResult, _ := C.recalibratePeaks(recalMeth, calibrantList,
		C.int(len(mzCalibrants)), C.int(*par.minCal), C.int(specNr))
	C.free(unsafe.Pointer(calibrantList))
	if int(specCalResult.satisfied) != 0 {
		for i := 0; i < int(specCalResult.nr_cal_pars); i++ {
			specRecalPar.P = append(specRecalPar.P, float64(specCalResult.cal_pars[i]))
		}
	}
	return specRecalPar, nil
}
