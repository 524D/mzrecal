#!/bin/bash
THISDIR=$PWD
DATA_DIR="${HOME}/data/mzrecal-application-note"
TOOLS_DIR="${HOME}/tools"
FN1='120118ry_201B7-32_2_2-120118ry007'
FN2='GSC11_24h_R1'
FASTA='UP000005640_9606.fasta'
T='/usr/bin/time -f %E'
COMET="${TOOLS_DIR}/comet.2019015.linux.exe"
# Comet expectation value limit that we accept
EXPECT=0.01

# Avoid locale errors in idconvert
export LC_ALL=C

# Extensions of files that we will create
# To avoid problems, we remove them before starting
RM_EXT='.pep.xml .mzid -recal.mzML -recal.mzid'

FN_BASE=${FN1}

# Convert raw data to mzML
# docker run -it --rm -e WINEDEBUG=-all -v ${DATA_DIR}:/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor" /data/${FN1}.wiff

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet "
echo $T "${COMET}" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet_${FN_BASE}.params" "${DATA_DIR}/${FN_BASE}.mzML"
$T "${COMET}" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet_${FN_BASE}.params" "${DATA_DIR}/${FN_BASE}.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Recalibration "
# $T "${TOOLS_DIR}/mzrecal" -empty-non-calibrated -mzTry=12 -mzAccept=2 -minPeak 500 -scoreFilter="MS:1002257(0.0:${EXPECT})" "${DATA_DIR}/${FN_BASE}.mzML"
$T "${TOOLS_DIR}/mzrecal" -mzTry=12 -mzAccept=2 -minPeak 500 -scoreFilter="MS:1002257(0.0:${EXPECT})" "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet on recalibrated output "
$T "${COMET}" -D${DATA_DIR}/${FASTA} "-P${DATA_DIR}/comet_${FN_BASE}.params" "${DATA_DIR}/${FN_BASE}-recal.mzML"

echo "Converting to .mzid  "
echo "${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.pep.xml"
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l

plot.R -e "${EXPECT}"  -m 12 "${DATA_DIR}/${FN_BASE}.mzid" "${DATA_DIR}/${FN_BASE}-recal.mzid"

FN_BASE=${FN2}

# Convert raw data to mzML
# docker run -it --rm -e WINEDEBUG=-all -v ${DATA_DIR}:/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor" /data/${FN1}.raw

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet "
$T "${COMET}" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet_${FN_BASE}.params" "${DATA_DIR}/${FN_BASE}.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Recalibration "
# $T "${TOOLS_DIR}/mzrecal" -empty-non-calibrated -mzTry=15 -mzAccept=2 -scoreFilter="MS:1002257(0.0:${EXPECT})" "${DATA_DIR}/${FN_BASE}.mzML"
$T "${TOOLS_DIR}/mzrecal" -mzTry=12 -mzAccept=2 -scoreFilter="MS:1002257(0.0:${EXPECT})" "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet on recalibrated output "
$T "${COMET}" -D${DATA_DIR}/${FASTA} "-P${DATA_DIR}/comet_${FN_BASE}.params" "${DATA_DIR}/${FN_BASE}-recal.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l

plot.R -e "${EXPECT}" -m 5 "${DATA_DIR}/${FN_BASE}.mzid" "${DATA_DIR}/${FN_BASE}-recal.mzid"

