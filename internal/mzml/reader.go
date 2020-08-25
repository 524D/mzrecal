package mzml

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"math"
	"strconv"

	"golang.org/x/net/html/charset"
)

// Read reads mzML file from an io.Reader
func Read(reader io.Reader) (MzML, error) {
	var mzML MzML

	d := xml.NewDecoder(reader)
	d.CharsetReader = charset.NewReaderLabel

	// We are only interested in mzML content, so skip over indexedmzML
	// and everything else
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			if tokenErr == io.EOF {
				break
			}
			return mzML, tokenErr
		}
		switch t := t.(type) {
		case xml.StartElement:
			if t.Name.Local == "mzML" {
				if err := d.DecodeElement(&mzML.content, &t); err != nil {
					if err != nil {
						return mzML, err
					}
				}
			}
		}
	}

	err := mzML.traverseScan()
	// Don't know why cleaning up the namespace is needed, but Go's XML
	// parser puts crap there
	//	mzML.cleanNamespaceAttrs()
	return mzML, err
}

// cleanNamespaceAttrs removes name space related attributes that are
// made up by Go's XML parser
// func (f *MzML) cleanNamespaceAttrs() {
// 	var newAttrs []xml.Attr
// 	// for i, attr := range f.content.Attrs {
// 	// 	if attr == xml.Attr{Name : "x", Value:"y"} {
// 	// 		newAtrrs = append(newAttrs, attr)
// 	// 	}
// 	// }
// 	f.content.Attrs = newAttrs
// }

// binaryDataPars decodes the CV terms in a mzML binarydata section
//
// CV Terms for binary data compression
// MS:1000574 zlib compression
// MS:1000576 No Compression
// MS:1002312 MS-Numpress linear prediction compression
// MS:1002313 MS-Numpress positive integer compression
// MS:1002314 MS-Numpress short logged float compression
// MS:1002746 MS-Numpress linear prediction compression followed by zlib compression
// MS:1002747 MS-Numpress positive integer compression followed by zlib compression
// MS:1002748 MS-Numpress short logged float compression followed by zlib compression
//
// CV Terms for binary data array types
// MS:1000514 m/z array
// MS:1000515 intensity array
//
// CV Terms for binary-data-type
// MS:1000521 32-bit float
// MS:1000523 64-bit float
func binaryDataPars(binaryDataArray *binaryDataArray) (
	bool, bool, bool, bool, error) {
	zlibCompression := bool(false) // Default: no compression
	bits64 := bool(false)          // Default: 32 bits
	mzArray := bool(false)
	intensityArray := bool(false)
	for _, cvParam := range binaryDataArray.CvPar {
		switch cvParam.Accession {
		case `MS:1000574`: // zlib compression
			zlibCompression = true
		case `MS:1000514`: // m/z array
			mzArray = true
		case `MS:1000515`: // intensity array
			intensityArray = true
		case `MS:1000523`: // 64-bit float
			bits64 = true
		case `MS:1002312`, `MS:1002313`, `MS:1002314`,
			`MS:1002746`, `MS:1002747`, `MS:1002748`:
			// MS-Numpress compression types
			log.Fatalf("Compression type not supported (CV term %s", cvParam.Accession)
		}
	}
	return zlibCompression, bits64, mzArray, intensityArray, nil
}

func fillScan(p []Peak, binaryDataArray *binaryDataArray) ([]Peak, error) {
	zlibCompression, bits64, mzArray, intensityArray, _ :=
		binaryDataPars(binaryDataArray)
	// We are only interrested in mz and intensity
	if mzArray || intensityArray {
		data, err := base64.StdEncoding.DecodeString(binaryDataArray.Binary)
		if err != nil {
			return nil, err
		}
		if zlibCompression {
			b := bytes.NewReader(data)
			z, err := zlib.NewReader(b)
			if err != nil {
				return nil, err
			}
			defer z.Close()
			d, err := ioutil.ReadAll(z)
			if err != nil {
				return nil, err
			}
			data = d
		}
		if bits64 {
			cnt := len(data) / 8
			if mzArray {
				for i := 0; i < cnt; i++ {
					bits := binary.LittleEndian.Uint64(data[i*8:])
					float := math.Float64frombits(bits)
					p[i].Mz = float64(float)
				}
			} else {
				for i := 0; i < cnt; i++ {
					bits := binary.LittleEndian.Uint64(data[i*8:])
					float := math.Float64frombits(bits)
					p[i].Intens = float64(float)
				}
			}
		} else {
			cnt := len(data) / 4
			if mzArray {
				for i := 0; i < cnt; i++ {
					bits := binary.LittleEndian.Uint32(data[i*4:])
					float := math.Float32frombits(bits)
					p[i].Mz = float64(float)
				}
			} else {
				for i := 0; i < cnt; i++ {
					bits := binary.LittleEndian.Uint32(data[i*4:])
					float := math.Float32frombits(bits)
					p[i].Intens = float64(float)
				}
			}
		}
	}
	return p, nil
}

// NumSpecs returns the number of spectra
func (f *MzML) NumSpecs() int {
	return len(f.content.Run.SpectrumList.Spectrum)
}

// RetentionTime returns the retention time of a spectrum
func (f *MzML) RetentionTime(scanIndex int) (float64, error) {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return 0.0, ErrInvalidScanIndex
	}
	for _, scan := range f.content.Run.SpectrumList.Spectrum[scanIndex].ScanList.Scan {
		for _, cvParam := range scan.CvPar {
			if cvParam.Accession == "MS:1000016" {
				retentionTime, err := strconv.ParseFloat(cvParam.Value, 64)
				// Check if the retention time is in minutes, otherwise assume it's seconds
				if cvParam.UnitAccession == "UO:0000031" ||
					cvParam.UnitAccession == "MS:1000038" {
					retentionTime *= 60
				}

				return retentionTime, err
			}
		}
	}
	return -1.0, nil
}

// IonInjectionTime returns the ion injection time of a spectrum in ms,
// or NaN is not found
func (f *MzML) IonInjectionTime(scanIndex int) (float64, error) {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return 0.0, ErrInvalidScanIndex
	}
	for _, scan := range f.content.Run.SpectrumList.Spectrum[scanIndex].ScanList.Scan {
		for _, cvParam := range scan.CvPar {
			if cvParam.Accession == "MS:1000927" {
				t, err := strconv.ParseFloat(cvParam.Value, 64)
				// Check if the ion injection time is in miliseconds,
				// (always the case currently), otherwise return error
				if cvParam.UnitAccession != "UO:0000028" {
					return t, ErrUnknownUnit
				}

				return t, err
			}
		}
	}
	return math.NaN(), nil
}

// ReadScan reads a single scan
// n is the sequence number of the scan in the mzML file,
// This is not the same as the scan number that is specified
// in the mzML file! To read a scan using the mzML number,
// use ReadScan(f, ScanIndex(f, scanNum))
func (f *MzML) ReadScan(scanIndex int) ([]Peak, error) {

	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return nil, ErrInvalidScanIndex
	}
	p := make([]Peak, f.content.Run.SpectrumList.Spectrum[scanIndex].DefaultArrayLength)
	var err error
	for _, b := range f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray {
		p, err = fillScan(p, &b)
		if err != nil {
			return p, err

		}
	}
	return p, nil
}

// Centroid returns true is the spectrum contains centroid peaks
func (f *MzML) Centroid(scanIndex int) (bool, error) {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return false, ErrInvalidScanIndex
	}

	for _, cvParam := range f.content.Run.SpectrumList.Spectrum[scanIndex].CvPar {
		if cvParam.Accession == "MS:1000127" { // centroid spectrum
			return true, nil
		}
	}
	return false, nil
}

// TotalIonCurrent returns the total ion current, or NaN if not found
func (f *MzML) TotalIonCurrent(scanIndex int) (float64, error) {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return 0.0, ErrInvalidScanIndex
	}

	for _, cvParam := range f.content.Run.SpectrumList.Spectrum[scanIndex].CvPar {
		if cvParam.Accession == "MS:1000285" { // total ion current
			tic, err := strconv.ParseFloat(cvParam.Value, 64)
			return tic, err
		}
	}
	return math.NaN(), nil
}

// MSLevel returns the MS level of a scan
func (f *MzML) MSLevel(scanIndex int) (int, error) {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return 0, ErrInvalidScanIndex
	}

	for _, cvParam := range f.content.Run.SpectrumList.Spectrum[scanIndex].CvPar {
		if cvParam.Accession == "MS:1000511" { // ms level
			msLevel, err := strconv.ParseInt(cvParam.Value, 10, 64)
			return int(msLevel), err
		}
	}
	return 1, nil // If nothing else, guess it's MS1
}

// MSInstruments returns the CV terms of the MS instrument
func (f *MzML) MSInstruments() ([]string, error) {

	type analyzer struct {
		CvPar CVParam `xml:"cvParam"`
	}
	type instrumentConfiguration struct {
		XMLName  xml.Name   `xml:"instrumentConfiguration"`
		Analyzer []analyzer `xml:"componentList>analyzer"`
	}

	var instr []string
	var instrConf instrumentConfiguration

	// Get the raw XML for the instrument configuration
	XML := f.content.InstrumentConfigurationList.InstrumentConfigurationListXML
	// Parse it
	err := xml.Unmarshal(XML, &instrConf)
	if err != nil {
		return nil, err
	}

	// Fill array with CV params of analysers
	for _, conf := range instrConf.Analyzer {
		instr = append(instr, conf.CvPar.Accession)
	}
	return instr, nil
}

// traverseScan traverses all scans,
// collects info of all scans and
// and fills the arrays f.index2id and f.id2Index to make scans accessible
func (f *MzML) traverseScan() error {

	f.index2id = make([]string, f.NumSpecs(), f.NumSpecs())
	f.id2Index = make(map[string]int, f.NumSpecs())
	err := error(nil)

	for i := range f.content.Run.SpectrumList.Spectrum {
		err = f.addSpecToIndex(i)
		if err != nil {
			return err
		}
	}
	return err
}

func (f *MzML) addSpecToIndex(i int) error {

	if i != f.content.Run.SpectrumList.Spectrum[i].Index {
		return ErrInvalidScanIndex
	}
	f.index2id[i] = f.content.Run.SpectrumList.Spectrum[i].ID
	f.id2Index[f.content.Run.SpectrumList.Spectrum[i].ID] = i
	return nil
}

// ScanIndex converts a scan identifier (the string used in the mzML file)
// into an index that is used to access the scans
func (f *MzML) ScanIndex(scanID string) (int, error) {
	if index, ok := f.id2Index[scanID]; ok {
		return index, nil
	}
	return 0, ErrInvalidScanID
}

// ScanID converts a scan index (used to access the scan data) into a scan id
// (used in the mzML file)
func (f *MzML) ScanID(scanIndex int) (string, error) {
	if scanIndex >= 0 && scanIndex < f.NumSpecs() {
		return f.index2id[scanIndex], nil
	}
	return "", ErrInvalidScanIndex
}

// GetPrecursors returns the mzML precursus struct for a given scanIndex
func (f *MzML) GetPrecursors(scanIndex int) ([]XMLprecursor, error) {
	if scanIndex >= 0 && scanIndex < f.NumSpecs() {
		var p []XMLprecursor
		if f.content.Run.SpectrumList.Spectrum[scanIndex].PrecursorList != nil {
			p = f.content.Run.SpectrumList.Spectrum[scanIndex].PrecursorList[0].Precursor
		}
		return p, nil
	}
	return nil, ErrInvalidScanIndex
}
