package mzidentml

import (
	"log"
	"os"
	"testing"
)

// Test files, downloaded Pride
const testFile1 = "/home/robm/data/mzid_testfiles/F181955 -filtered.mzid"
const testFile2 = "/home/robm/data/mzid_testfiles/BalbfedbyRag_Day7_01.mzid"

// const testFile = "/home/robm/data/mzid_testfiles/BalbfedbyRag_Day7_01.mzid"

func TestAll1(t *testing.T) {

	x, err := os.Open(testFile1)
	if err != nil {
		t.Errorf("Open: mzIdentMLfile is nil")
	}
	f, err := Read(x)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	//	log.Printf("%+v\n", f.content)

	n := f.NumIdents()
	if n != 8600 {
		t.Errorf("NumIdents is %d, expected 8600", n)
	}
	ident, err := f.Ident(1000)
	if err != nil {
		t.Errorf("Ident: error return %v", err)
	}
	log.Printf("ident: %+v", ident)

}

func TestAll2(t *testing.T) {

	x, err := os.Open(testFile2)
	if err != nil {
		t.Errorf("Open: mzIdentMLfile is nil")
	}
	f, err := Read(x)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	//	log.Printf("%+v\n", f.content)

	n := f.NumIdents()
	if n != 8754 {
		t.Errorf("NumIdents is %d, expected 8754", n)
	}
	ident, err := f.Ident(1000)
	if err != nil {
		t.Errorf("Ident: error return %v", err)
	}
	log.Printf("ident: %+v", ident)

}
