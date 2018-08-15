package mzml

import (
	"encoding/xml"
	"errors"
)

type MzML struct {
	content  mzMLContent
	index2id []string
	id2Index map[string]int
}

// CV Terms for binary data compression
// MS:1000574 zlib compression
// MS:1000576 No Compression
// MS:1002312 MS-Numpress linear prediction compression
// MS:1002313 MS-Numpress positive integer compression
// MS:1002314 MS-Numpress short logged float compression
// MS:1002746 MS-Numpress linear prediction compression followed by zlib compression
// MS:1002747 MS-Numpress positive integer compression followed by zlib compression
// MS:1002748 MS-Numpress short logged float compression followed by zlib compression

// CV Terms for binary data array types
// MS:1000514 m/z array
// MS:1000515 intensity array

// CV Terms for binary-data-type
// MS:1000521 32-bit float
// MS:1000523 64-bit float

// Peak contains the actual ms peak info
type Peak struct {
	Mz     float64
	Intens float64
}

// The mzML content that we read. Not all fields are parsed,
// but we need to store them in order to write the result mzML.
type mzMLContent struct {
	XMLName         xml.Name `xml:"mzML"`
	CvList          cvList   `xml:"cvList"`
	FileDescription struct {
		FileDescriptionXML string `xml:",innerxml"`
	} `xml:"fileDescription"`
	ReferenceableParamGroupList *referenceableParamGroupList `xml:"referenceableParamGroupList,omitempty"`
	SoftwareList                *softwareList                `xml:"softwareList,omitempty"`
	InstrumentConfigurationList *instrumentConfigurationList `xml:"instrumentConfigurationList,omitempty"`
	DataProcessingList          *dataProcessingList          `xml:"dataProcessingList,omitempty"`
	Spectrum                    []spectrum                   `xml:"run>spectrumList>spectrum,omitempty"`
}

type cvList struct {
	Count     int    `xml:"count,attr"`
	CvListXML []byte `xml:",innerxml"`
}

type referenceableParamGroupList struct {
	Count                          int    `xml:"count,attr"`
	ReferenceableParamGroupListXML []byte `xml:",innerxml"`
}

type softwareList struct {
	Count           int    `xml:"count,attr"`
	SoftwareListXML []byte `xml:",innerxml"`
}

type instrumentConfigurationList struct {
	Count                          int    `xml:"count,attr"`
	InstrumentConfigurationListXML []byte `xml:",innerxml"`
}

type dataProcessingList struct {
	Count                 int    `xml:"count,attr"`
	DataProcessingListXML []byte `xml:",innerxml"`
}

type spectrum struct {
	ID                 string            `xml:"id,attr"`
	DefaultArrayLength int64             `xml:"defaultArrayLength,attr"`
	Index              int               `xml:"index,attr"`
	CvParam            []cvParam         `xml:"cvParam"`
	ScanCvParam        []cvParam         `xml:"scanList>scan>cvParam"`
	BinaryDataArray    []binaryDataArray `xml:"binaryDataArrayList>binaryDataArray"`
}

type cvParam struct {
	Accession     string `xml:"accession,attr"`
	Name          string `xml:"name,attr"`
	Value         string `xml:"value,attr"`
	UnitAccession string `xml:"unitAccession,attr"`
}

type binaryDataArray struct {
	CvParam []cvParam `xml:"cvParam"`
	Binary  string    `xml:"binary"`
}

var (
	ErrInvalidScanId    = errors.New("MzML: invalid scan id")
	ErrInvalidScanIndex = errors.New("MzML: invalid scan index")
)
