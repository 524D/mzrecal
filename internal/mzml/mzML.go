package mzml

import (
	"encoding/xml"
	"errors"
)

// MzML wraps the contents of the mzML file
type MzML struct {
	content  mzMLContent
	index2id []string
	id2Index map[string]int
}

// type Precursor struct {
// 	spectrumRef string
// 	selectedIon []Peak
// }

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
	XMLName         xml.Name `xml:"http://psi.hupo.org/ms/mzml mzML"`
	CvList          cvList   `xml:"cvList"`
	FileDescription struct {
		FileDescriptionXML string `xml:",innerxml"`
	} `xml:"fileDescription"`
	ReferenceableParamGroupList *referenceableParamGroupList `xml:"referenceableParamGroupList"`
	SoftwareList                *softwareList                `xml:"softwareList"`
	InstrumentConfigurationList *instrumentConfigurationList `xml:"instrumentConfigurationList"`
	DataProcessingList          *dataProcessingList          `xml:"dataProcessingList"`
	Run                         run                          `xml:"run"`
}

// We define a separte struct for writing XML because it is not possible
// to write namespace info otherwise
type mzMLContentWrite struct {
	XMLName         xml.Name `xml:"http://psi.hupo.org/ms/mzml mzML"`
	Sl1             string   `xml:"xsi:schemaLocation,attr"`
	Version         string   `xml:"version,attr"`
	Sl2             string   `xml:"xmlns:xsi,attr"`
	CvList          cvList   `xml:"cvList"`
	FileDescription struct {
		FileDescriptionXML string `xml:",innerxml"`
	} `xml:"fileDescription"`
	ReferenceableParamGroupList *referenceableParamGroupList `xml:"referenceableParamGroupList,omitempty"`
	SoftwareList                *softwareList                `xml:"softwareList"`
	InstrumentConfigurationList *instrumentConfigurationList `xml:"instrumentConfigurationList"`
	DataProcessingList          *dataProcessingList          `xml:"dataProcessingList"`
	Run                         run                          `xml:"run"`
}

type cvList struct {
	Count     int    `xml:"count,attr,omitempty"`
	CvListXML []byte `xml:",innerxml"`
}

type referenceableParamGroupList struct {
	Count                          int    `xml:"count,attr,omitempty"`
	ReferenceableParamGroupListXML []byte `xml:",innerxml"`
}

type softwareList struct {
	Count    int        `xml:"count,attr,omitempty"`
	Software []software `xml:"software"`
}

type software struct {
	ID      string    `xml:"id,attr,omitempty"`
	Version string    `xml:"version,attr,omitempty"`
	CvPar   []CVParam `xml:"cvParam,omitempty"`
}

type instrumentConfigurationList struct {
	Count                          int    `xml:"count,attr,omitempty"`
	InstrumentConfigurationListXML []byte `xml:",innerxml"`
}

type dataProcessingList struct {
	Count           int              `xml:"count,attr,omitempty"`
	DataProcessingd []DataProcessing `xml:"dataProcessing,omitempty"`
}

type DataProcessing struct {
	ID             string             `xml:"id,attr,omitempty"`
	ProcessingMeth []ProcessingMethod `xml:"processingMethod"`
}

type ProcessingMethod struct {
	Count       int         `xml:"order,attr,omitempty"`
	SoftwareRef string      `xml:"softwareRef,attr,omitempty"`
	CvPar       []CVParam   `xml:"cvParam,omitempty"`
	UserPar     []userParam `xml:"userParam,omitempty"`
}

type run struct {
	Id                                string           `xml:"id,attr,omitempty"`
	DefaultInstrumentConfigurationRef string           `xml:"defaultInstrumentConfigurationRef,attr,omitempty"`
	StartTimeStamp                    string           `xml:"startTimeStamp,attr,omitempty"`
	DefaultSourceFileRef              string           `xml:"defaultSourceFileRef,attr,omitempty"`
	SpectrumList                      spectrumList     `xml:"spectrumList,omitempty"`
	ChromatogramList                  chromatogramList `xml:"chromatogramList,omitempty"`
}

type spectrumList struct {
	Count                    int        `xml:"count,attr,omitempty"`
	DefaultDataProcessingRef string     `xml:"defaultDataProcessingRef,attr,omitempty"`
	Spectrum                 []spectrum `xml:"spectrum,omitempty"`
}

type chromatogramList struct {
	Count                    int    `xml:"count,attr,omitempty"`
	DefaultDataProcessingRef string `xml:"defaultDataProcessingRef,attr,omitempty"`
	ChromatogramListXML      []byte `xml:",innerxml"`
}

type spectrum struct {
	Index              int       `xml:"index,attr"`
	ID                 string    `xml:"id,attr"`
	DefaultArrayLength int64     `xml:"defaultArrayLength,attr"`
	CvPar              []CVParam `xml:"cvParam,omitempty"`
	ScanList           scanList  `xml:"scanList"`
	// precursorList is a slice, only the current version of
	// the encoding/xml package does not handle "omitempty" properly on
	// structures, and we don't want precursorList tags to appear in
	// e.g. ms1 spectra
	PrecursorList       []precursorList     `xml:"precursorList,omitempty"`
	BinaryDataArrayList binaryDataArrayList `xml:"binaryDataArrayList"`
}

type binaryDataArrayList struct {
	Count           int               `xml:"count,attr,omitempty"`
	BinaryDataArray []binaryDataArray `xml:"binaryDataArray"`
}

type binaryDataArray struct {
	EncodedLength int       `xml:"encodedLength,attr,omitempty"`
	ArrayLength   int       `xml:"arrayLength,attr,omitempty"`
	CvPar         []CVParam `xml:"cvParam,omitempty"`
	Binary        string    `xml:"binary"`
}

type scanList struct {
	Count int       `xml:"count,attr,omitempty"`
	CvPar []CVParam `xml:"cvParam,omitempty"`
	Scan  []scan    `xml:"scan"`
}

type scan struct {
	InstrConfRef   string         `xml:"instrumentConfigurationRef,attr,omitempty"`
	CvPar          []CVParam      `xml:"cvParam,omitempty"`
	UserPar        []userParam    `xml:"userParam,omitempty"`
	ScanWindowList scanWindowList `xml:"scanWindowList"`
}

type userParam struct {
	Name  string `xml:"name,attr,omitempty"`
	Value string `xml:"value,attr,omitempty"`
	Type  string `xml:"type,attr,omitempty"`
}

type precursorList struct {
	Count     int            `xml:"count,attr,omitempty"`
	Precursor []XMLprecursor `xml:"precursor"`
}

type XMLprecursor struct {
	SpectrumRef     string          `xml:"spectrumRef,attr,omitempty"`
	IsolationWindow isolationWindow `xml:"isolationWindow,omitempty"`
	SelectedIonList selectedIonList `xml:"selectedIonList"`
	Activation      activation      `xml:"activation"`
}

type isolationWindow struct {
	CvPar []CVParam `xml:"cvParam,omitempty"`
}

type selectedIonList struct {
	Count       int           `xml:"count,attr,omitempty"`
	CvPar       []CVParam     `xml:"cvParam,omitempty"`
	SelectedIon []selectedIon `xml:"selectedIon"`
}

type selectedIon struct {
	CvPar []CVParam `xml:"cvParam,omitempty"`
}

type activation struct {
	CvPar []CVParam `xml:"cvParam,omitempty"`
}

type scanWindowList struct {
	Count          int    `xml:"count,attr,omitempty"`
	ScanWindowList string `xml:",innerxml"`
}

// CVParam contains values and attributes of a mzML Controlled Vocabulary term
// (http://www.peptideatlas.org/tmp/mzML1.1.0.html)
type CVParam struct {
	Accession     string `xml:"accession,attr,omitempty"`
	Name          string `xml:"name,attr,omitempty"`
	Value         string `xml:"value,attr,omitempty"`
	UnitCvRef     string `xml:"unitCvRef,attr,omitempty"`
	UnitAccession string `xml:"unitAccession,attr,omitempty"`
	UnitName      string `xml:"unitName,attr,omitempty"`
}

var (
	ErrInvalidScanId    = errors.New("MzML: invalid scan id")
	ErrInvalidScanIndex = errors.New("MzML: invalid scan index")
	ErrUnknownUnit      = errors.New("MzML: can't handle unit")
)
