package mzml

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"io"
	"math"
)

func (f *MzML) Write(writer io.Writer) error {
	writer.Write(([]byte)(
		`<?xml version="1.0" encoding="utf-8"?>
`))
	enc := xml.NewEncoder(writer)
	// FIXME: We want readable XML, with XML tags starting on a new line.
	// GO's Encode doesn't always insert newlines, and using
	// Indent only works if the indent string is not empty,
	// resuling in a single space indent.
	enc.Indent(` `, `  `)
	var content mzMLContentWrite

	content.XMLName = f.content.XMLName
	content.Sl1 = "http://psi.hupo.org/ms/mzml http://psidev.info/files/ms/mzML/xsd/mzML1.1.0.xsd"
	content.Version = "1.1.0"
	content.Sl2 = "http://www.w3.org/2001/XMLSchema-instance"
	content.CvList = f.content.CvList
	content.FileDescription = f.content.FileDescription
	content.ReferenceableParamGroupList = f.content.ReferenceableParamGroupList
	content.SoftwareList = f.content.SoftwareList
	content.InstrumentConfigurationList = f.content.InstrumentConfigurationList
	content.DataProcessingList = f.content.DataProcessingList
	content.Run = f.content.Run

	err := enc.Encode(&content)
	return err
}

// AppendSoftwareInfo adds info to the SoftwareList tag of the mzML file
func (f *MzML) AppendSoftwareInfo(id string, version string) error {
	var sw software

	sw.ID = id
	sw.Version = version
	f.content.SoftwareList.Count++
	f.content.SoftwareList.Software = append(f.content.SoftwareList.Software, sw)
	return nil
}

// AppendDataProcessing adds info to the DataProcessing tag of the mzML file
func (f *MzML) AppendDataProcessing(proc DataProcessing) error {
	f.content.DataProcessingList.Count++
	f.content.DataProcessingList.DataProcessingd = append(f.content.DataProcessingList.DataProcessingd, proc)
	return nil
}

// UpdateScan sets the mz/intensity info of a scan
func (f *MzML) UpdateScan(scanIndex int, p []Peak,
	updateMz bool, updateIntens bool) error {
	if scanIndex < 0 || scanIndex >= f.NumSpecs() {
		return ErrInvalidScanIndex
	}
	// Workaround for msConvert:
	// Insert a dummy peak if there is none, otherwise msConvert generates an error
	if len(p) == 0 {
		var peak Peak
		p = append(p, peak)
	}

	f.content.Run.SpectrumList.Spectrum[scanIndex].DefaultArrayLength = int64(len(p))
	for i := range f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray {
		zlibCompression, bits64, mzArray, intensityArray, err :=
			binaryDataPars(&f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray[i])
		if err != nil {
			return err
		}
		// We are only interested in mz and intensity
		if (mzArray && updateMz) || (intensityArray && updateIntens) {

			b64, err := encodeBinary(p, zlibCompression, bits64, mzArray)
			if err != nil {
				return err
			}
			f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray[i].Binary = b64
			f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray[i].ArrayLength = len(p)
			f.content.Run.SpectrumList.Spectrum[scanIndex].BinaryDataArrayList.BinaryDataArray[i].EncodedLength = len(b64)
		}
	}
	return nil
}

func encodeBinary(p []Peak, zlibCompression bool, bits64 bool, mzArray bool) (
	string, error) {

	var data []byte
	var rawUncompressed []byte

	// Some code duplication below in order to optimize loops
	if bits64 {
		// Allocate room for uncompressed binary data
		rawUncompressed = make([]byte, len(p)*8)
		if mzArray {
			for i, peak := range p {
				u64bits := math.Float64bits(peak.Mz)
				binary.LittleEndian.PutUint64(rawUncompressed[(8*i):], u64bits)
			}
		} else {
			for i, peak := range p {
				u64bits := math.Float64bits(peak.Intens)
				binary.LittleEndian.PutUint64(rawUncompressed[(8*i):], u64bits)
			}
		}
	} else {
		rawUncompressed = make([]byte, len(p)*4)
		if mzArray {
			for i, peak := range p {
				u32bits := math.Float32bits(float32(peak.Mz))
				binary.LittleEndian.PutUint32(rawUncompressed[(4*i):], u32bits)
			}
		} else {
			for i, peak := range p {
				u32bits := math.Float32bits(float32(peak.Intens))
				binary.LittleEndian.PutUint32(rawUncompressed[(4*i):], u32bits)
			}
		}
	}
	if zlibCompression {
		var b bytes.Buffer
		z := zlib.NewWriter(&b)
		defer z.Close()
		z.Write(rawUncompressed)
		z.Close() // zlib writer must explicitly be closed here, otherwise resu is invalid
		data = b.Bytes()
	} else {
		data = rawUncompressed
	}
	encodedStr := base64.StdEncoding.EncodeToString(data)
	return encodedStr, nil
}
