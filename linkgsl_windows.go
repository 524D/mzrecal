package main

// This file is only compiled for Windows
// Static link gsl library

/*
#cgo LDFLAGS: -Lwindows/gsl/lib/
#cgo LDFLAGS: -l:libgsl.a
#cgo LDFLAGS: -l:libgslcblas.a
#cgo LDFLAGS: -lm
*/
import "C"
