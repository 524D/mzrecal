package main

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
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

// struct for URL, filename, and boolean for whether the file is gzipped
type testFile struct {
	url      string
	filename string
	gzipped  bool
}

// struct for list of test files, and destination directory
type testFilesUrlName struct {
	files []testFile
	dir   string
}

// The files that we want to download for testing
var testFiles = testFilesUrlName{
	files: []testFile{
		{"https://osf.io/download/hjk8z/", "test.mzML", false},
		{"https://osf.io/download/8agvn/", "test.mzid", false},
		{"https://osf.io/download/v7hrf/", "test-recal_base.mzML", false},
		{"https://osf.io/download/wgtnf/", "test-recal_base.json", false},
	},
	dir: "testdata",
}

// Download gets a file from a given URL, and puts it in the supplied directory
// If isGzip is true, the file is assumed to be gzip compressed
// and is uncompressed before writing to disk
func downloadFile(url string, dir string, filename string, isGzip bool) error {

	// The final output file name
	fn := filepath.Join(dir, filename)
	// Create a temporary file for download
	tmpFile, err := os.CreateTemp(dir, filename)
	if err != nil {
		return err
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	r := resp.Body
	// Uncompress if needed
	if isGzip {
		r, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer r.Close()
	}
	_, err = io.Copy(tmpFile, r)
	if err != nil {
		return err
	}
	tmpFile.Close()
	os.Rename(tmpFile.Name(), fn)
	return nil
}

func ensureTestData(t testing.TB) {
	// Ensure the directory from testFilesToDownload exists
	if _, err := os.Stat(testFiles.dir); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(testFiles.dir, 0755)
		if err != nil {
			t.Fatalf("Error creating test data directory: %v", err)
		}
	}

	// Download the test files from testFilesToDownload if they don't exist
	for _, f := range testFiles.files {
		fullPath := filepath.Join(testFiles.dir, f.filename)
		if _, err := os.Stat(fullPath); errors.Is(err, os.ErrNotExist) {
			err := downloadFile(f.url, testFiles.dir, f.filename, f.gzipped)
			if err != nil {
				t.Fatalf("Error downloading test file: %v", err)
			}
		}
	}
}

func JSONCompare(t testing.TB, expected, actual io.Reader) {
	alwaysEqual := cmp.Comparer(func(_, _ interface{}) bool { return true })

	opts := cmp.Options{
		// This option declares that a float64 comparison is equal only if
		// both inputs are NaN.
		cmp.FilterValues(func(x, y float64) bool {
			return math.IsNaN(x) && math.IsNaN(y)
		}, alwaysEqual),

		// This option declares approximate equality on float64s only if
		// both inputs are not NaN.
		cmp.FilterValues(func(x, y float64) bool {
			return !math.IsNaN(x) && !math.IsNaN(y)
		}, cmp.Comparer(func(x, y float64) bool {
			delta := math.Abs(x - y)
			mean := math.Abs(x+y) / 2.0
			return delta/mean < 0.00001
		})),
	}

	var in1 map[string]any
	var in2 map[string]any

	dec := json.NewDecoder(expected)
	err := dec.Decode(&in1)
	if err != nil {
		t.Fatalf("Error decoding expected JSON: %v", err)
	}
	dec = json.NewDecoder(actual)
	err = dec.Decode(&in2)
	if err != nil {
		t.Fatalf("Error decoding actual JSON: %v", err)
	}

	if diff := cmp.Diff(in1, in2, opts); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}

}

// JSONCompareFile compares the contents of two JSON files
func JSONCompareFile(t testing.TB, expectedFile, actualFile string) {
	expected, err := os.Open(expectedFile)
	if err != nil {
		t.Fatalf("Error opening expected file: %v", err)
	}
	defer expected.Close()
	actual, err := os.Open(actualFile)
	if err != nil {
		t.Fatalf("Error opening actual file: %v", err)
	}
	defer actual.Close()
	JSONCompare(t, expected, actual)
}

func TestMain(t *testing.T) {
	ensureTestData(t)

	// // Test case 1: No arguments
	// os.Args = []string{"mzrecal"}
	// main()

	// // Test case 2: Help argument
	// os.Args = []string{"mzrecal", "-h"}
	// main()

	// Test case 3: Using test data, auto-generate mzid file name
	os.Args = []string{"mzrecal", filepath.Join(testFiles.dir, testFiles.files[0].filename)}
	main()
	JSONCompareFile(t, filepath.Join(testFiles.dir, testFiles.files[3].filename), filepath.Join(testFiles.dir, "test-recal.json"))
}
