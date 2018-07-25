package main

// Dynamic link against gsl
// #cgo LDFLAGS: -lgsl -lgslcblas -lm

// Static link against gls

// #cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgsl.a
// #cgo LDFLAGS: /usr/lib/x86_64-linux-gnu/libgslcblas.a
// #cgo LDFLAGS: -lm
// int add1(int i)
// { return (i+1); }
// #include <gsl/gsl_qrng.h>
import "C"
import "fmt"

func gsl() {
	var rndVals [10]C.double

	i := C.add1(1)
	fmt.Println("1+1=", i)
	rg := C.gsl_qrng_alloc(C.gsl_qrng_niederreiter_2, 10)
	for i := 0; i < 100; i++ {
		C.gsl_qrng_get(rg, &rndVals[0])
		fmt.Println("rndVals=", rndVals)
		r0 := float64(rndVals[0])
		fmt.Println("r0:", r0)
	}
	C.gsl_qrng_free(rg)
}
