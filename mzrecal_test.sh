#!/bin/bash
TARGET_DIR='/home/robm/mzrecal'
DATA_DIR='/home/robm/data/msrecal_ribosomes'
TOOLS_DIR='/home/robm/tools'
FN_BASE='human_ribosome_60S_bottomup_peak'
FASTA='uniprothuman_20180620.fasta'
T='/usr/bin/time -f %E'

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
$T "${TOOLS_DIR}/comet.2018012.linux.exe" "-D${DATA_DIR}/${FASTA}" "-P${TOOLS_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}.mzML" >/dev/null

echo "Converting to .pep.xml "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Computing recalibration "
$T "${TARGET_DIR}/mzrecal" -scoreFilter="MS:1002257(0.0:0.05)" "${DATA_DIR}/${FN_BASE}.mzML"
echo -n "Creating recalibrated output "
$T "${TARGET_DIR}/mzrecal" -recal "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running msconvert (generating indexed mzml) "
$T "${TOOLS_DIR}/msconvert" "${DATA_DIR}/${FN_BASE}-recal.mzML" --outfile "${FN_BASE}-recal.indexed.mzML" -o "${DATA_DIR}"  >/dev/null
echo -n "Running comet on recalibrated output "
$T "${TOOLS_DIR}/comet.2018012.linux.exe" -D${DATA_DIR}/${FASTA} "-P${TOOLS_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}-recal.indexed.mzML"  >/dev/null

echo "Converting to .pep.xml "
"$TOOLS_DIR/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.indexed.pep.xml" >/dev/null 2>/dev/null
# Note: idconvert does not handle files with multiple extensions correctly,
# hence the .indexed extension is missing
echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l

# Older tests
#go run mzrecal.go gsl_wrapper.go  --mzml="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzML" --mzid="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzid"
#go run mzrecal.go gsl_wrapper.go --scoreFilter="MS:1002257(0.0:1e0)" --mzid="${DATA_DIR}/${FN_BASE}.mzid" --cal="${DATA_DIR}/${FN_BASE}-recal.json" "${DATA_DIR}/${FN_BASE}.mzML"

# Run test from docker container
# docker run -v /home/robm/data:/data  mzrecal /mzrecal  --scoreFilter="MS:1002257(0.0:1e0)" --mzml="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML" --mzid="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" --cal="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak-recal.json"
