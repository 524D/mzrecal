package main

// This file is only compiled for Linux arm
// Static link gsl library

/*
#cgo LDFLAGS: -L/usr/lib/arm-linux-gnueabihf/
#cgo LDFLAGS: -l:libgsl.a
#cgo LDFLAGS: -l:libgslcblas.a
#cgo LDFLAGS: -lm
*/
import "C"
