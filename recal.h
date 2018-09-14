#ifndef RECAL_H
#define RECAL_H

// WARNING: the following enum must stay consistent with the iota in mzrecal.go
typedef enum {
    CALIB_NONE,
    CALIB_FTICR,
    CALIB_TOF,
    CALIB_ORBITRAP,
} calib_method_t;

// The maximum number of calibration parameters of any calibration function
#define MAX_CAL_PARS 10

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

double mz_recalX(double mz_meas, cal_params_t *p);
cal_params_t recalibratePeaks(recal_data_t d,
                              int min_cal,
                              double internal_calibration_target,
                              int spec_nr);
void fill_calibrant_list(calibrant_t *calibrant_list, int i,
                         double mz_calc, double mz_measured);

#endif
