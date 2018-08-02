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
// typedef struct calibrant_type {
//     double mz; /* calculated m/z */
//     double peak; /* measured m/z */
//     double intensity;
// } calibrant;
//
//
// struct data {
//     double * y; /* measured f (in phony units) */
//     double * mz2; /* theoretical m/z */
// };
//
// // FIXME: Global data and fixed size arrays: This sucks!
// #define MAX_CALIBRANTS 10000
// #define INTERNAL_CALIBRATION_TARGET 3e-6	//discard internal calibrants that do not fit CAL2 better than this
// calibrant calibrant_list[MAX_CALIBRANTS];
// calibrant calibrant_list_tmp[MAX_CALIBRANTS];
// int n_calibrants;
//
// double mz_recal(double peak, double Ca, double Cb)
// {
// 	return Ca/((1/peak)-Cb);
// }
//
// int calib_f(const gsl_vector *x, void *params, gsl_vector *f)
// {
// 	double *y = ((struct data *)params)->y;
// 	double *mz = ((struct data *)params)->mz2;
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
// 	double *y = ((struct data *)params)->y;
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
// spec_cal_result_t recalibratePeaks(int min_cal, int spec_nr){
//     int status, SATISFIED, j;
//
//     const gsl_multifit_fdfsolver_type *T;
// 	   gsl_multifit_fdfsolver *s;
// 	   double chi;
//     int iter=0;
// 	   const size_t pp=2; /* number of free parameters in calibration function */
// 	   double y[MAX_CALIBRANTS];
// 	   double mz2[MAX_CALIBRANTS];
// 	   struct data d={y,mz2};
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
//             d.y[j] = 1 / calibrant_list[j].peak;
//             d.mz2[j] = calibrant_list[j].mz;
//         }
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
//         int j_tmp = 0;
//         for(j=0; j<n_calibrants; j++) {
//             if (fabs((calibrant_list[j].mz-mz_recal(calibrant_list[j].peak, Ca, Cb))/calibrant_list[j].mz)<INTERNAL_CALIBRATION_TARGET) {
//                 calibrant_list_tmp[j_tmp++] = calibrant_list[j];
//             }
//         }
//         for(j=0; j<j_tmp; j++) {
//             calibrant_list[j] = calibrant_list_tmp[j];
//         }
//         if (j_tmp == n_calibrants) {
//             SATISFIED=1; /* all calibrants < INTERNAL_CALIBRATION_TARGET */
//         }
//         n_calibrants=j_tmp;
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
import "C"

func recalibrateSpec(specNr int, mzCalibrants []mzCalibrant, par params) (specRecalParams, error) {
	var specRecalPar specRecalParams

	specRecalPar.SpecNr = specNr

	for i, calibrant := range mzCalibrants {
		C.calibrant_list[i].mz = C.double(calibrant.mz)
		C.calibrant_list[i].peak = C.double(calibrant.mzMeasured)
	}
	C.n_calibrants = C.int(len(mzCalibrants))
	specCalResult, _ := C.recalibratePeaks(C.int(*par.minCal), C.int(specNr))
	if int(specCalResult.satisfied) != 0 {
		for i := 0; i < int(specCalResult.nr_cal_pars); i++ {
			specRecalPar.P = append(specRecalPar.P, float64(specCalResult.cal_pars[i]))
		}
	}
	return specRecalPar, nil
}
