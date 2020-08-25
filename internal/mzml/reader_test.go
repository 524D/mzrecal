package mzml

import (
	"log"
	"math"
	"os"
	"testing"
)

// Test files, downloaded from Pride
const testFile1 = "/home/robm/data/mzml_testfiles/100_mzML/100.mzML"
const testFile2 = "/home/robm/data/mzml_testfiles/NJ-ManIxCC124-6-SN-FASP-v6-ISF80-12052016.mzML"
const testFile3 = "/home/robm/data/mzml_testfiles/PT4708-8.mzML"

func TestAll1(t *testing.T) {
	x, err := os.Open(testFile1)
	if err != nil {
		t.Errorf("Open: mzMLfile is nil")
	}
	f, err := Read(x)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	} else {
		//	log.Print("f.content.DataProcessingList :", f.content.DataProcessingList, "Count: ", f.content.DataProcessingList.Count)
		p, err := f.ReadScan(0)
		if err != nil {
			t.Errorf("ReadScan: error return %v", err)
		}
		if p[0].Mz < 699.695 || p[0].Mz > 699.696 {
			t.Errorf("ReadScan: peak 0 mz %v", p[0].Mz)
		}
		// log.Printf("%+v", p)
		centroid, err := f.Centroid(0)
		if err != nil {
			t.Errorf("Centroid: error return %v", err)
		}
		if centroid {
			t.Errorf("Centroid: true, should be false")
		}
		_, err = f.Centroid(1)
		if err != ErrInvalidScanIndex {
			t.Errorf("Centroid: error return %v, should be ErrInvalidScanIndex", err)
		}
		msLevel, err := f.MSLevel(0)
		if err != nil {
			t.Errorf("MSLevel: error return %v", err)
		}
		if msLevel != 1 {
			t.Errorf("MSLevel: %d, should be 1", msLevel)
		}
		_, err = f.MSLevel(1)
		if err != ErrInvalidScanIndex {
			t.Errorf("MSLevel: error return %v, should be ErrInvalidScanIndex", err)
		}

		_, err = f.ScanIndex(`file=sourceFile1`)
		if err != ErrInvalidScanID {
			t.Errorf("ScanIndex: error return %v, should be ErrInvalidScanID", err)
		}
		scanIndex, err := f.ScanIndex(`file=sourceFile`)
		if err != nil {
			t.Errorf("ScanIndex: error return %v", err)
		}
		if scanIndex != 0 {
			t.Errorf("ScanIndex: %d, should be 0", scanIndex)
		}
		_, err = f.ScanID(1)
		if err != ErrInvalidScanIndex {
			t.Errorf("ScanIndex: error return %v, should be ErrInvalidScanID", err)
		}
		scanID, err := f.ScanID(0)
		if err != nil {
			t.Errorf("ScanID: error return %v", err)
		}
		if scanID != `file=sourceFile` {
			t.Errorf("ScanID: %s, should be file=sourceFile", scanID)
		}
	}
	n := f.NumSpecs()
	if n != 1 {
		t.Errorf("NumSpecs: %d, should be 1", n)
	}

	instruments, err := f.MSInstruments()
	if err != nil {
		t.Errorf("MSInstruments: error return %v", err)
	} else {
		log.Printf("reader_test Instrument CV term: %+v\n", instruments)
	}

}

func TestAll2(t *testing.T) {
	x, err := os.Open(testFile2)
	if err != nil {
		t.Errorf("Open: mzMLfile is nil")
	}
	f, err := Read(x)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	// log.Printf("%+v", f.content)

	p, err := f.ReadScan(0)
	if err != nil {
		t.Errorf("ReadScan: error return %v", err)
	}
	if p[0].Mz < 601.744 || p[0].Mz > 601.745 {
		t.Errorf("ReadScan: peak 0 mz %v", p[0].Mz)
	}
	// log.Printf("%+v", p)
	centroid, err := f.Centroid(10)
	if err != nil {
		t.Errorf("Centroid: error return %v", err)
	}
	if !centroid {
		t.Errorf("Centroid: true, should be true")
	}
	_, err = f.Centroid(-1)
	if err != ErrInvalidScanIndex {
		t.Errorf("Centroid: error return %v, should be ErrInvalidScanIndex", err)
	}
	msLevel, err := f.MSLevel(0)
	if err != nil {
		t.Errorf("MSLevel: error return %v", err)
	}
	if msLevel != 1 {
		t.Errorf("MSLevel: %d, should be 1", msLevel)
	}
	_, err = f.MSLevel(-123)
	if err != ErrInvalidScanIndex {
		t.Errorf("MSLevel: error return %v, should be ErrInvalidScanIndex", err)
	}

	_, err = f.ScanIndex(`blabla`)
	if err != ErrInvalidScanID {
		t.Errorf("ScanIndex: error return %v, should be ErrInvalidScanID", err)
	}
	scanIndex, err := f.ScanIndex(`controllerType=0 controllerNumber=1 scan=63`)
	if err != nil {
		t.Errorf("ScanIndex: error return %v", err)
	}
	if scanIndex != 62 {
		t.Errorf("ScanIndex: %d, should be 62", scanIndex)
	}

	_, err = f.ScanID(666666)
	if err != ErrInvalidScanIndex {
		t.Errorf("ScanIndex: error return %v, should be ErrInvalidScanID", err)
	}
	scanID, err := f.ScanID(42)
	if err != nil {
		t.Errorf("ScanID: error return %v", err)
	}
	if scanID != `controllerType=0 controllerNumber=1 scan=43` {
		t.Errorf("ScanID: %s, should be file=sourceFile", scanID)
	}

	n := f.NumSpecs()
	if n != 11873 {
		t.Errorf("NumSpecs: %d, should be 11873", n)
	}
	rt, _ := f.RetentionTime(100)
	if err != nil {
		t.Errorf("RetentionTime: error return %v", err)
	}
	expectRt := float64(326.59077)
	if math.Abs((rt/expectRt)-1.0) > 0.000001 {
		t.Errorf("RetentionTime: %f, should be %f", rt, expectRt)
	}

}
