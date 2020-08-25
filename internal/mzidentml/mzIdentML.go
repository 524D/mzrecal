package mzidentml

import (
	"encoding/xml"
	"errors"
)

// Types for parsing mzIdentML

// MzIdentML holds only the part of mzIdentML files
// in which we are interrested
type MzIdentML struct {
	seqID2PepIdx map[string]int
	identList    []identRef
	content      mzIdentMLContent
}

type identRef struct {
	specResultIdx int // Index into SpectrumIdentificationResult
	specIDIdx     int // Index into SpectrumIdentificationItem

}

type Identification struct {
	PepSeq        string
	PepID         string
	Charge        int
	ModMass       float64
	SpecID        string
	RetentionTime float64
	Cv            []cvParam
}

type mzIdentMLContent struct {
	XMLName                      xml.Name                       `xml:"MzIdentML"`
	Peptide                      []peptide                      `xml:"SequenceCollection>Peptide"`
	SpectrumIdentificationResult []spectrumIdentificationResult `xml:"DataCollection>AnalysisData>SpectrumIdentificationList>SpectrumIdentificationResult"`
}

type peptide struct {
	ID              string `xml:"id,attr"`
	PeptideSequence string
	Modification    []modification
}

type modification struct {
	// Note: monoisotopicMassDelta is optional according the the schema, but
	// appears to be no other way to determine mass shift, as other
	// corresponding cvParam's don't carry this info either
	MonoisotopicMassDelta float64 `xml:"monoisotopicMassDelta,attr"`
}

type spectrumIdentificationResult struct {
	SpectrumID                 string `xml:"spectrumID,attr"`
	SpectrumIdentificationItem []spectrumIdentificationItem
	CvPar                      []cvParam `xml:"cvParam"`
}

type spectrumIdentificationItem struct {
	ChargeState int       `xml:"chargeState,attr"`
	PeptideRef  string    `xml:"peptide_ref,attr"`
	CvPar       []cvParam `xml:"cvParam"`
}

type cvParam struct {
	Accession     string `xml:"accession,attr"`
	Name          string `xml:"name,attr"`
	Value         string `xml:"value,attr"`
	UnitAccession string `xml:"unitAccession,attr"`
}

var (
	ErrInvalidIdentIndex = errors.New("mzIdentML: invalid identification index")
)
