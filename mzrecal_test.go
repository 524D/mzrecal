package main

import (
	"errors"
	"testing"
)

func TestParseFloat64Range(t *testing.T) {
	// Test case 1: Valid input range
	min, max, err := parseFloat64Range("0.5:1.5", 0.0, 2.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != 0.5 {
		t.Errorf("Expected min to be 0.5, got: %f", min)
	}
	if max != 1.5 {
		t.Errorf("Expected max to be 1.5, got: %f", max)
	}

	// Test case 2: Empty input range
	min, max, err = parseFloat64Range("", 0.0, 2.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != 0.0 {
		t.Errorf("Expected min to be 0.0, got: %f", min)
	}
	if max != 2.0 {
		t.Errorf("Expected max to be 2.0, got: %f", max)
	}

	// Test case 3: Invalid input range
	min, max, err = parseFloat64Range("2.5:1.5", 0.0, 2.0)
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	if !errors.Is(err, ErrRangeSpec) {
		t.Errorf("Expected error: %v, got: %v", ErrRangeSpec, err)
	}
	if min != 1.5 {
		t.Errorf("Expected min to be 1.5, got: %f", min)
	}
	if max != 1.5 {
		t.Errorf("Expected max to be 1.5, got: %f", max)
	}

	// Test case 4: Only max specified
	min, max, err = parseFloat64Range(":1.5", 0.0, 2.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != 0.0 {
		t.Errorf("Expected min to be 0.0, got: %f", min)
	}
	if max != 1.5 {
		t.Errorf("Expected max to be 1.5, got: %f", max)
	}

	// Test case 5: Only min specified
	min, max, err = parseFloat64Range("0.5:", 0.0, 2.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != 0.5 {
		t.Errorf("Expected min to be 0.5, got: %f", min)
	}
	if max != 2.0 {
		t.Errorf("Expected max to be 2.0, got: %f", max)
	}

	// Test case 6: Only ":" specified
	min, max, err = parseFloat64Range(":", 0.0, 2.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != 0.0 {
		t.Errorf("Expected min to be 0.0, got: %f", min)
	}
	if max != 2.0 {
		t.Errorf("Expected max to be 2.0, got: %f", max)
	}

	// Test case 7: Exponents in numbers
	min, max, err = parseFloat64Range("-2.0e10:3.0e10", -1000000000000.0, 1000000000000.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != -2.0e10 {
		t.Errorf("Expected min to be -2.0e10, got: %f", min)
	}
	if max != 3.0e10 {
		t.Errorf("Expected max to be 3.0e10, got: %f", max)
	}

	// Test case 8: Out of range
	min, max, err = parseFloat64Range("-2.0:2.0", -1.0, 1.0)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if min != -1.0 {
		t.Errorf("Expected min to be -1.0, got: %f", min)
	}
	if max != 1.0 {
		t.Errorf("Expected max to be 1.0, got: %f", max)
	}
}
