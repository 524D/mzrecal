package main

// This file is only compiled for Linux x86_64
// Static link gsl library

/*
#cgo LDFLAGS: -L/usr/lib/x86_64-linux-gnu/
#cgo LDFLAGS: -l:libgsl.a
#cgo LDFLAGS: -l:libgslcblas.a
#cgo LDFLAGS: -lm
*/
import "C"
