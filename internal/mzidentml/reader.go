package mzidentml

import (
	"encoding/xml"
	"io"
	"math"
	"strconv"

	"golang.org/x/net/html/charset"
)

// Read reads mzIdentML content from io.reader
func Read(reader io.Reader) (MzIdentML, error) {
	var mzIdentML MzIdentML
	d := xml.NewDecoder(reader)
	d.CharsetReader = charset.NewReaderLabel
	err := d.Decode(&mzIdentML.content)
	if err != nil {
		return mzIdentML, err
	}
	mzIdentML.buildPepID2Sequence()
	mzIdentML.buildIdentList()
	return mzIdentML, err
}

func (m *MzIdentML) buildPepID2Sequence() {
	m.seqID2PepIdx = make(map[string]int, len(m.content.Peptide))
	for i, p := range m.content.Peptide {
		m.seqID2PepIdx[p.ID] = i
	}
}

func (m *MzIdentML) buildIdentList() {
	for i := range m.content.SpectrumIdentificationResult {
		for j := range m.content.SpectrumIdentificationResult[i].SpectrumIdentificationItem {
			var iRef identRef
			iRef.specIDIdx = i
			iRef.specResultIdx = j
			m.identList = append(m.identList, iRef)
		}
	}
}

// NumIdents returns the total number of identifications in the mzIdentML file
// Note that for some spectra, multiple identifications may be present
// The identifications can be accessed using the Ident() method, which takes
// an index as argument. The index runs from 0 to NumIdents()-1
func (m *MzIdentML) NumIdents() int {
	return len(m.identList)
}

// Ident returns a spectrum identification from the mzIdentML file.
// Parameter i is the index of the identification to return. The index runs
// from 0 to NumIdents()-1
func (m *MzIdentML) Ident(i int) (Identification, error) {

	var ident Identification

	if i < 0 || i >= len(m.identList) {
		return ident, ErrInvalidIdentIndex
	}
	specIDIdx := m.identList[i].specIDIdx
	specResultIdx := m.identList[i].specResultIdx

	pepRef := m.content.SpectrumIdentificationResult[specIDIdx].SpectrumIdentificationItem[specResultIdx].PeptideRef
	pepIdx := m.seqID2PepIdx[pepRef]
	ident.PepSeq = m.content.Peptide[pepIdx].PeptideSequence
	ident.PepID = m.content.Peptide[pepIdx].ID
	ident.ModMass = float64(0)
	ident.Charge = m.content.SpectrumIdentificationResult[specIDIdx].SpectrumIdentificationItem[specResultIdx].ChargeState
	for _, mod := range m.content.Peptide[pepIdx].Modification {
		ident.ModMass += mod.MonoisotopicMassDelta
	}
	ident.SpecID = m.content.SpectrumIdentificationResult[specIDIdx].SpectrumID
	ident.RetentionTime = float64(-1)
	prio := math.MaxInt32
	for _, cv := range m.content.SpectrumIdentificationResult[specIDIdx].CvPar {
		// There are multiple CV terms that can be used to report the
		// retention time. In order of decreasing preference we use:
		// 1. MS:1000016 - scan start time
		// 2. MS:1000894 - retention time
		// 3. MS:1000826 - elution time
		// 4. MS:1001114 - retention time (deprecated)
		useTime := false
		switch cv.Accession {
		case "MS:1000016":
			if prio > 1 {
				prio = 1
				useTime = true
			}
		case "MS:1000894":
			if prio > 2 {
				prio = 2
				useTime = true
			}
		case "MS:1000826":
			if prio > 3 {
				prio = 3
				useTime = true
			}
		case "MS:1001114":
			if prio > 4 {
				prio = 4
				useTime = true
			}
		}
		// If a (higher priority) term was found, process/store the retention time
		if useTime {
			retentionTime, err := strconv.ParseFloat(cv.Value, 64)
			if err != nil {
				return ident, err
			}
			// Check if the retention time is in minutes, otherwise assume it's seconds
			if cv.UnitAccession == "UO:0000031" || cv.UnitAccession == "MS:1000038" {
				retentionTime *= 60
			}
			ident.RetentionTime = retentionTime
		}
	}
	// Collect CV terms/values for the identification, the scores are in there
	for _, cv := range m.content.SpectrumIdentificationResult[specIDIdx].SpectrumIdentificationItem[specResultIdx].CvPar {
		ident.Cv = append(ident.Cv, cv)
	}

	return ident, nil
}
