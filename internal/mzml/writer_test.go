package mzml

import (
	"os"
	"testing"
)

const testFileWriteIn = "/home/robm/data/mzml_testfiles/NJ-ManIxCC124-6-SN-FASP-v6-ISF80-12052016.mzML"
const testFile1Write = "/home/robm/data/mzml_testfiles/write_1.mzML"
const testFile1Writea = "/home/robm/data/mzml_testfiles/write_1a.mzML"
const testFile1Writeb = "/home/robm/data/mzml_testfiles/write_1b.mzML"

func TestAllWrite1(t *testing.T) {
	rf, err := os.Open(testFileWriteIn)
	if err != nil {
		t.Errorf("Open %s error: %v", testFile1, err)
	}
	f, err := Read(rf)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	p, err := f.ReadScan(0)
	if err != nil {
		t.Errorf("ReadScan: error return %v", err)
	}
	rf.Close()
	p[0].Mz = 42.0
	p[0].Intens = 777.0
	f.UpdateScan(0, p, true, false)
	p[0].Mz = 42.0
	p[0].Intens = 777.0
	f.UpdateScan(1, p, false, true)
	wf, err := os.Create(testFile1Write)
	if err != nil {
		t.Errorf("Create %s error: %v", testFile1Write, err)
	}
	err = f.Write(wf)
	//	fmt.Printf("%+v\n", f.content.Attrs)
	if err != nil {
		t.Errorf("Write: error return %v", err)
	}
	wf.Close()
	rf, err = os.Open(testFile1Write)
	if err != nil {
		t.Errorf("Open %s error: %v", testFile1Write, err)
	}
	f, err = Read(rf)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	p, err = f.ReadScan(0)
	if err != nil {
		t.Errorf("ReadScan: error return %v", err)
	}
	// Check if only mz changed for scan 0 peak 0
	if p[0].Mz < 41.9999 || p[0].Mz > 42.0001 {
		t.Errorf("ReadScan: peak 0 mz %v", p[0].Mz)
	}
	if p[0].Intens > 776.9999 && p[0].Intens < 777.0001 {
		t.Errorf("ReadScan: peak 0 intens %v", p[0].Intens)
	}
	p, err = f.ReadScan(1)
	if err != nil {
		t.Errorf("ReadScan: error return %v", err)
	}
	// Check if only intens changed for scan 1 peak 0
	if p[0].Mz > 41.9999 && p[0].Mz < 42.0001 {
		t.Errorf("ReadScan: peak 1 mz %v", p[0].Mz)
	}
	if p[0].Intens < 776.9999 || p[0].Intens > 777.0001 {
		t.Errorf("ReadScan: peak 1 intens %v", p[0].Intens)
	}
	rf.Close()

	// Check if after some iterations, the file stays the same
	wf, err = os.Create(testFile1Writea)
	if err != nil {
		t.Errorf("Create %s error: %v", testFile1Writea, err)
	}
	err = f.Write(wf)
	if err != nil {
		t.Errorf("Write: error return %v", err)
	}
	wf.Close()
	rf, err = os.Open(testFile1Writea)
	if err != nil {
		t.Errorf("Open %s error: %v", testFile1Writea, err)
	}
	f, err = Read(rf)
	if err != nil {
		t.Errorf("Read: error return %v", err)
	}
	rf.Close()
	wf, err = os.Create(testFile1Writeb)
	if err != nil {
		t.Errorf("Create %s error: %v", testFile1Writeb, err)
	}
	err = f.Write(wf)
	if err != nil {
		t.Errorf("Write: error return %v", err)
	}
	wf.Close()
	// // Compute SHA256 over files to see if they are equal
	// hf1, err := os.Open("testFile1Writea")
	// if err != nil {
	// 	t.Errorf("Open %s error: %v", testFile1Writea, err)
	// }
	// //	defer hf1.Close()
	// h1 := sha256.New()
	// // Why doesn't this work??
	// if _, err := io.Copy(h1, hf1); err != nil {
	// 	fmt.Printf("a\n")
	// 	log.Fatal(err)
	// }
	// hf1.Close()
	// fmt.Printf("b\n")
	// hf2, err := os.Open("testFile1Writeb")
	// if err != nil {
	// 	t.Errorf("Open %s error: %v", testFile1Writeb, err)
	// }
	// defer hf2.Close()
	// h2 := sha256.New()
	// if _, err := io.Copy(h2, hf2); err != nil {
	// 	log.Fatal(err)
	// }
	// if string(h1.Sum(nil)) != string(h2.Sum(nil)) {
	// 	t.Errorf("Checksum different after 2 consecutive read/writes")
	//
	// }
}
