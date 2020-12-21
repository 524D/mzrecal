// Based on MSRECAL/RECAL_FUNCTIONS.C
// We use qsort_r, which is a GNU extension
#define _GNU_SOURCE

#include <stdlib.h>
#include <stdio.h>
#include <gsl/gsl_vector.h>
#include <gsl/gsl_blas.h>
#include <gsl/gsl_multifit_nlin.h>
#include "recal.h"

// Maximum number of iterations for the FDF solver
#define MAX_FDF_SOLVER_ITER 500
#define EPS_ABS 1e-9
#define EPS_REL 1e-9

static double mz_recal_poly_n(double mz_meas, cal_params_t *p, int degree)
{
  int i;
  double mp = 1.0;
  double mz_calib = 0.0;
  for (i=0; i<=degree; i++) {
      mz_calib += p->cal_pars[i]*mp;
      mp *= mz_meas;
  }
  return mz_calib;
}

double mz_recalX(double mz_meas, cal_params_t *p)
{
    double mz_calib;
    switch (p->calib_method) {
    case CALIB_FTICR:
        // mz_calib = Ca/((1/mz_meas)-Cb)
        mz_calib = (p->cal_pars[1])/((1/mz_meas)-(p->cal_pars[0]));
        break;
    case CALIB_TOF:
        mz_calib = p->cal_pars[2]*sqrt(mz_meas) + p->cal_pars[1]*mz_meas + p->cal_pars[0];
        break;
    case CALIB_ORBITRAP: {
        // mz_calib = A/((f-B)^2) =
        //      A / ((1/sqrt(mz_meas))-B)^2
        double a = p->cal_pars[1];
        double b = p->cal_pars[0];
        double freq;
        double fb;

        freq = 1/sqrt(mz_meas);
        fb = freq - b;
        mz_calib = a/(fb*fb);
        }
        break;
    case CALIB_OFFSET:
        mz_calib = mz_meas + p->cal_pars[0];
        break;
    case CALIB_POLY1:
        mz_calib = mz_recal_poly_n(mz_meas, p, 1);
        break;
    case CALIB_POLY2:
        mz_calib = mz_recal_poly_n(mz_meas, p, 2);
        break;
    case CALIB_POLY3:
        mz_calib = mz_recal_poly_n(mz_meas, p, 3);
        break;
    case CALIB_POLY4:
        mz_calib = mz_recal_poly_n(mz_meas, p, 4);
        break;
    case CALIB_POLY5:
        mz_calib = mz_recal_poly_n(mz_meas, p, 5);
        break;
    default:
        mz_calib = mz_meas;
        break;
    }
    return mz_calib;
}

static void calib_f_poly_n(const gsl_vector *x, recal_data_t *recal_data_p,
   gsl_vector *f, int degree)
{
  int i, j;
  double mz_pow, mz_meas, mz_calib, cal_pars[MAX_CAL_PARS];
  calibrant_t *calibrants = recal_data_p->calibrants;
  int n_calibrants = recal_data_p->n_calibrants;

  for (j=0; j<=degree; j++) {
      cal_pars[j] = gsl_vector_get (x, j);
  }

  for (i=0;i<n_calibrants;i++) {
    mz_pow = 1.0;
    mz_calib = 0.0;
    mz_meas = calibrants[i].mz_meas;
    for (j=0; j<=degree; j++) {
        mz_calib += cal_pars[j]*mz_pow;
        mz_pow *= mz_meas;
    }
    gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
  }
}

static int calib_f(const gsl_vector *x, void *params, gsl_vector *f)
{
    recal_data_t *recal_data_p = (recal_data_t *)params;

    calibrant_t *calibrants = recal_data_p->calibrants;
    int n_calibrants = recal_data_p->n_calibrants;
    int rv = GSL_SUCCESS;
    size_t i;

    switch (recal_data_p->calib_method) {
    case CALIB_FTICR: {
        double a = gsl_vector_get (x, 1);
        double b = gsl_vector_get (x, 0);
        double mz_calib;
        double freq;

        for (i=0;i<n_calibrants;i++) {
            // Model m = a/(f-b) (CAL2 inverted)
            freq= 1/calibrants[i].mz_meas;
            mz_calib = a/(freq-b);
            gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
        }
        break;
    }
    case CALIB_TOF: {
        double a = gsl_vector_get (x, 2);
        double b = gsl_vector_get (x, 1);
        double c = gsl_vector_get (x, 0);
        double mz_calib;
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            mz_calib = a*sqrt(mz_meas) + b*mz_meas + c;
            gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
        }
        break;
    }
    case CALIB_ORBITRAP: {
        double a = gsl_vector_get (x, 1);
        double b = gsl_vector_get (x, 0);
        double mz_calib;
        double freq;
        double fb;

        for (i=0;i<n_calibrants;i++) {
            // mz_calib = A/((f-B)^2) =
            //      A / ((1/sqrt(mz_meas))-B)^2
            freq = 1/sqrt(calibrants[i].mz_meas);
            fb = freq - b;
            if (fb == 0.0) {
                rv = GSL_ERANGE;
            }
            else {
                mz_calib = a/(fb*fb);
                gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
            }
        }
        break;
    }
    case CALIB_OFFSET: {
        double a = gsl_vector_get (x, 0);
        double mz_calib;
        for (i=0;i<n_calibrants;i++) {
            // mz_calib = mz_meas + a
            mz_calib = calibrants[i].mz_meas + a;
            gsl_vector_set (f, i, (mz_calib-calibrants[i].mz_calc)); // absolute or relative error?
            }
        }
        break;
    case CALIB_POLY1:
        calib_f_poly_n(x, recal_data_p, f, 1);
        break;
    case CALIB_POLY2:
        calib_f_poly_n(x, recal_data_p, f, 2);
        break;
    case CALIB_POLY3:
        calib_f_poly_n(x, recal_data_p, f, 3);
        break;
    case CALIB_POLY4:
        calib_f_poly_n(x, recal_data_p, f, 4);
        break;
    case CALIB_POLY5:
        calib_f_poly_n(x, recal_data_p, f, 5);
        break;
     default:
         break;
     }
     return rv;
}

// DF calibrator
static int calib_df(const gsl_vector *x, void *params, gsl_matrix *J)
{
    recal_data_t *recal_data_p = (recal_data_t *)params;

    calibrant_t *calibrants = recal_data_p->calibrants;
    int n_calibrants = recal_data_p->n_calibrants;
    int rv = GSL_SUCCESS;
    size_t i;

    switch (recal_data_p->calib_method) {
    case CALIB_FTICR: {
        double a = gsl_vector_get (x, 1);
        double b = gsl_vector_get (x, 0);
        double freq;

        for (i=0;i<n_calibrants;i++) {
            freq=1/calibrants[i].mz_meas;
            gsl_matrix_set (J,i,0, a/((freq-b)*(freq-b)) );
            gsl_matrix_set (J,i,1, 1/(freq-b) );
        }
        break;
    }
    case CALIB_TOF:
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, mz_meas);
            gsl_matrix_set(J, i, 2, sqrt(mz_meas));
        }
        break;
    case CALIB_ORBITRAP:
        for (i=0;i<n_calibrants;i++) {
            double a = gsl_vector_get (x, 1);
            double b = gsl_vector_get (x, 0);
            double freq;

            for (i=0;i<n_calibrants;i++) {
                freq=1/sqrt(calibrants[i].mz_meas);
                if ((freq-b) == 0.0) {
                    rv = GSL_ERANGE;
                }
                else {
                    gsl_matrix_set (J,i,0, 2.0*a/((freq-b)*(freq-b)*(freq-b)));
                    gsl_matrix_set (J,i,1, 1.0/((freq-b)*(freq-b)));
                }
            }
        }
        break;
    case CALIB_OFFSET:
        for (i=0;i<n_calibrants;i++) {
            gsl_matrix_set(J, i, 0, 1);
        }
        break;
    case CALIB_POLY1:
        for (i=0;i<n_calibrants;i++) {
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, calibrants[i].mz_meas);
        }
        break;
    case CALIB_POLY2:
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, mz_meas);
            gsl_matrix_set(J, i, 2, mz_meas * mz_meas);
        }
        break;
    case CALIB_POLY3:
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            double m_pow = mz_meas;
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 2, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 3, m_pow);
        }
     break;
    case CALIB_POLY4:
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            double m_pow = mz_meas;
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 2, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 3, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 4, m_pow);
        }
        break;
    case CALIB_POLY5:
        for (i=0;i<n_calibrants;i++) {
            double mz_meas = calibrants[i].mz_meas;
            double m_pow = mz_meas;
            gsl_matrix_set(J, i, 0, 1);
            gsl_matrix_set(J, i, 1, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 2, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 3, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 4, m_pow);
            m_pow *= mz_meas;
            gsl_matrix_set(J, i, 5, m_pow);
        }
        break;
    default:
        break;
    }
    return rv;
}

// FDF Calibrator
static int calib_fdf(const gsl_vector *x, void *params, gsl_vector *f, gsl_matrix *J)
{
     int rv = calib_f(x, params, f);
     if (rv == GSL_SUCCESS) {
         rv = calib_df(x, params, J);
     }
     return rv;
}

int get_nr_cal_pars(calib_method_t calib_method) {
  int nr_cal_pars=0;
  switch (calib_method) {
  case CALIB_FTICR:
      nr_cal_pars = 2;
      break;
  case CALIB_TOF:
      nr_cal_pars = 3;
      break;
  case CALIB_ORBITRAP:
      nr_cal_pars = 2;
      break;
  case CALIB_OFFSET:
      nr_cal_pars = 1;
      break;
  case CALIB_POLY1:
      nr_cal_pars = 2;
      break;
  case CALIB_POLY2:
      nr_cal_pars = 3;
      break;
  case CALIB_POLY3:
      nr_cal_pars = 4;
      break;
  case CALIB_POLY4:
      nr_cal_pars = 5;
      break;
  case CALIB_POLY5:
      nr_cal_pars = 6;
      break;
  }
  return nr_cal_pars;
}

static void init_cal_params(cal_params_t *cal_params, calib_method_t calib_method) {
    int i;

    cal_params->calib_method = calib_method;
    // Initialize calibration parameters close to the optimum
    // For simplicity, init all values to zero
    for (i=0; i<MAX_CAL_PARS; i++) {
        cal_params->cal_pars[i] = 0.0;
    }
    cal_params->cal_pars[1] = 1.0;
    cal_params->nr_cal_pars = get_nr_cal_pars(calib_method);
}

// recal_err_rel computes the relative mass error of a calibrant
// after recalibration
static double recal_err_rel(calibrant_t *calibrant,
                            cal_params_t *cal_params) {
    return (calibrant->mz_calc-mz_recalX(calibrant->mz_meas, cal_params))/
            calibrant->mz_calc;
}

// compare_mzrecal_err compared (in qsort sense) the relative mass error
// of 2 calibrants after recalibration
#if (defined _WIN32 || defined _WIN64 || defined __WINDOWS__) 
static int compare_mzrecal_err(void *arg, const void *i, const void *j) {
#else
static int compare_mzrecal_err(const void *i, const void *j, void *arg) {
#endif
    cal_params_t *cal_params = (cal_params_t *)arg;

	double erri = recal_err_rel((calibrant_t *)i, cal_params);
	double errj = recal_err_rel((calibrant_t *)j, cal_params);

	if ((errj - erri) > 0)
		return -1;
	else if ((errj - erri) < 0)
		return 1;
	return 0;
}

// remove_outliers_mzQC removes outliers according to the definition of mzQC
// (https://github.com/HUPO-PSI/mzQC/blob/bulk-cvterms/cv/qc-cv.obo)
static int remove_outliers_mzQC(recal_data_t *d, cal_params_t *cal_params, int debug) {
    int satisfied = 0;
    // Sort calibrants by error
#if (defined _WIN32 || defined _WIN64 || defined __WINDOWS__) 
    qsort_s(d->calibrants, d->n_calibrants, sizeof(calibrant_t), compare_mzrecal_err, cal_params);
#else
    qsort_r(d->calibrants, d->n_calibrants, sizeof(calibrant_t), compare_mzrecal_err, cal_params);
#endif

    // mzQC definition of outliers used Q1, Q3 and IQR of the distribution,
    // compute them here
    int q1_i1, q1_i2;
    // Special case: for < 6 calibrants, adapt method for mzQC to work well
    if (d->n_calibrants < 6) {
        // For less than 4 calibrants, we omit outlier detection
        if (d->n_calibrants < 4) {
            satisfied = 1;
            return 1;
        }
        // For 4 to 5 calibrants, we use Q1 and Q3 are based on values 1 position from extreme
        else {
            q1_i1 = 1; 
            q1_i2 = 1;
        }
    }
    else {
        int nq1 = d->n_calibrants/2; // count of samples that Q1 is based on (odd numbers are rounded down)
        q1_i1 = (nq1-1)/2; // index 1 of the median of upper half
        q1_i2 = nq1/2; // index 2 of the median of upper half (for odd number of samples in upper half)
    }
    double q1 = (recal_err_rel(&(d->calibrants[q1_i1]), cal_params) +
            recal_err_rel(&(d->calibrants[q1_i2]), cal_params))/2;
    int q3_i1 = d->n_calibrants - q1_i1 -1;
    int q3_i2 = d->n_calibrants - q1_i2 -1;
    double q3 = (recal_err_rel(&(d->calibrants[q3_i1]), cal_params) +
            recal_err_rel(&(d->calibrants[q3_i2]), cal_params))/2;
    double iqr = q3-q1;
    // Compute outlier limits according to mzQC definition
    double ol_low_lim = q1 - 1.5 * iqr;
    double ol_high_lim = q3 + 1.5 * iqr;
    // Remove outliers
    // and recalibrate (as long as we have at least min_cal peaks)
    int accepted_idx = 0;
    for(int j=0; j<d->n_calibrants; j++) {
        double rel_err = recal_err_rel(&(d->calibrants[j]), cal_params);
        if (debug) {
            char s[3] = "  \0";
            if ( (j==q1_i1) || (j==q3_i1) ) {
                s[0] = '*';
            }
            if ( (j==q1_i2) || (j==q3_i2) ) {
                s[1] = '*';
            }
            printf("%d rel_err=%e %s\n", j, rel_err, s);
        }
        if ( (rel_err>=ol_low_lim) && (rel_err<=ol_high_lim)) {
            d->calibrants[accepted_idx++] = d->calibrants[j];
        }
    }
    if (debug) {
        printf("q1_i1=%d q1_i2=%d q3_i1=%d q3_i2=%d q1=%e q3=%e iqr=%e ol_low_lim=%e ol_high_lim=%e\n",
               q1_i1, q1_i2, q3_i1, q3_i2, q1, q3, iqr, ol_low_lim, ol_high_lim);
    }

    // If all (remaining) calibrants are accepted, we are done
    if (accepted_idx == d->n_calibrants) {
        satisfied=1; // all calibrants < internal_calibration_target
    }
    d->n_calibrants=accepted_idx;
    return satisfied;
}

cal_params_t recalibratePeaks(recal_data_t *d,
                              int min_cal,
                              double internal_calibration_target,
                              int spec_nr,
                              int debug){
    int status, satisfied, j, vi;

    const gsl_multifit_fdfsolver_type *T;
    gsl_multifit_fdfsolver *s;
    double chi;
    int iter=0;
    cal_params_t cal_params;

    if (debug) setvbuf(stdout, NULL, _IONBF, 0);  // Autoflush
    if (debug) printf("Spec %d\n", spec_nr);
    init_cal_params(&cal_params, d->calib_method);

    gsl_multifit_function_fdf func;

    func.f = &calib_f;
    func.df = &calib_df;
    func.fdf = &calib_fdf;

    satisfied=0;
    while (d->n_calibrants >= min_cal && !satisfied) {
        // ??? The array fed to gsl_vector_view_array needs to be a copy
        // otherwise the result is not the same
        cal_params_t cal_params_copy = cal_params;
        gsl_vector_view x=gsl_vector_view_array(cal_params_copy.cal_pars,cal_params.nr_cal_pars);
        // least-squares fit first using all peaks, than removing those that don't fit
        iter=0;
        T = gsl_multifit_fdfsolver_lmder;
        s = gsl_multifit_fdfsolver_alloc (T, d->n_calibrants, cal_params.nr_cal_pars);

        func.n = d->n_calibrants;
        func.p = cal_params.nr_cal_pars;
        func.params = d;
        gsl_multifit_fdfsolver_set(s,&func,&x.vector);

        do {
            iter++;
            status = gsl_multifit_fdfsolver_iterate (s);

            if (status) {
                if (debug) printf("gsl_multifit_fdfsolver_iterate status %d, breaking loop\n", status);
                break;
            }
            status=gsl_multifit_test_delta (s->dx, s->x, EPS_ABS, EPS_REL);
        } while (status==GSL_CONTINUE && iter<MAX_FDF_SOLVER_ITER);
        if (debug) printf("gsl_multifit_test_delta status %d, iter %d \n", status, iter);

        for (vi=0; vi<cal_params.nr_cal_pars; vi++) {
            cal_params.cal_pars[vi] = gsl_vector_get(s->x,vi);
            if (debug) printf("cal_pars[%d]=%f\n", vi, cal_params.cal_pars[vi]);
        }

        chi = gsl_blas_dnrm2(s->f);
        gsl_multifit_fdfsolver_free(s);

        // remove outliers

        // If no explicit max error is specified, we use mzQC definition of outlier
        // (https://github.com/HUPO-PSI/mzQC/blob/bulk-cvterms/cv/qc-cv.obo)
        if (internal_calibration_target == 0.0) {
            satisfied = remove_outliers_mzQC(d, &cal_params, debug);
        }
        else {
            int accepted_idx = 0;
            // check if all calibrants are < internal_calibration_target, if not, throw these out
            // and recalibrate (as long as we have at least min_cal peaks)
            for(j=0; j<d->n_calibrants; j++) {
                double mz_calc = d->calibrants[j].mz_calc;
                double mz_meas = d->calibrants[j].mz_meas;
                double mz_recal = mz_recalX(mz_meas, &cal_params);
                if (debug)
                    printf("mz_calc=%f mz_meas=%f mz_recal=%f errppm=%f accept=%d\n", mz_calc, mz_meas, mz_recal,
                    1000000*((mz_calc-mz_recal)/mz_calc), (int)(fabs((mz_calc-mz_recal)/mz_calc)<internal_calibration_target));
                if (fabs((mz_calc-mz_recal)/mz_calc)<internal_calibration_target) {
                    d->calibrants[accepted_idx++] = d->calibrants[j];
                }
            }
            // If all (remaining) calibrants are accepted, we are done
            if (accepted_idx == d->n_calibrants) {
                satisfied=1; // all calibrants < internal_calibration_target
            }
            d->n_calibrants=accepted_idx;
        }
    }
    cal_params.n_calibrants = d->n_calibrants;
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
    calibrant_list[i].id = i;
}

// Function get_calibrant_id is only needed because directly accessing
// a C "pointer array" from Go is a bit messy.
int get_calibrant_id(calibrant_t *calibrant_list, int i) {
  return calibrant_list[i].id;
}
