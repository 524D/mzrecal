#!/bin/bash
TARGET_DIR="${HOME}/mzrecal"
DATA_DIR="${HOME}/data/msrecal_ribosomes"
TOOLS_DIR="${HOME}/tools"
FN_BASE='human_ribosome_60S_bottomup_peak'
FASTA='uniprothuman_20180620.fasta'
T='/usr/bin/time -f %E'
COMET="${TOOLS_DIR}/comet.2019011.linux.exe"

# Extensions of files that we will create
# To avoid problems, we remove them before starting
RM_EXT='.pep.xml .mzid -recal.mzML -recal.indexed.mzML -recal.indexed.pep.xml -recal.mzid'

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"
echo -n "Building mzrecal "
mkdir -p "${TARGET_DIR}"
$T go build -a -ldflags '-extldflags "-static"' -o "${TARGET_DIR}/mzrecal"

echo -n "Running comet "
$T "${COMET}/comet.2018014.linux.exe" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}.mzML" >/dev/null

echo "Converting to .pep.xml "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Recalibration "
$T "${TARGET_DIR}/mzrecal" -scorefilter="MS:1002257(0.0:0.05)" "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running msconvert (generating indexed mzml) "
$T "${TOOLS_DIR}/msconvert" "${DATA_DIR}/${FN_BASE}-recal.mzML" --outfile "${FN_BASE}-recal.indexed.mzML" -o "${DATA_DIR}"  >/dev/null
echo -n "Running comet on recalibrated output "
$T "${COMET}" -D${DATA_DIR}/${FASTA} "-P${DATA_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}-recal.indexed.mzML"  >/dev/null

echo "Converting to .pep.xml "
"$TOOLS_DIR/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.indexed.pep.xml" >/dev/null 2>/dev/null
# Note: idconvert does not handle files with multiple extensions correctly,
# hence the .indexed extension is missing
echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l

