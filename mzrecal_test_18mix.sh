#!/bin/bash
TARGET_DIR='/home/robm/mzrecal'
DATA_DIR='/home/robm/data/18_Mix'
TOOLS_DIR='/home/robm/tools'
FN_BASES='OR20070924_S_mix7_02 OR20070924_S_mix7_03 OR20070924_S_mix7_04 OR20070924_S_mix7_05 OR20070924_S_mix7_02 OR20070924_S_mix7_06 OR20070924_S_mix7_07 OR20070924_S_mix7_08 OR20070924_S_mix7_09 OR20070924_S_mix7_10 OR20070924_S_mix7_11'
FASTA='18mix_db_plus_contaminants_20081209.fasta'
T='/usr/bin/time -f %E'

# Extensions of files that we will create
# To avoid problems, we remove them before starting
RM_EXT='.pep.xml .mzid .recal.mzML .recal.indexed.mzML .recal.indexed.pep.xml .recal.mzid'

for FN_BASE in ${FN_BASES}
do

XINTERACT_FILES_UNCALIBRATED="${XINTERACT_FILES_UNCALIBRATED} ${FN_BASE}.pep.xml"
XINTERACT_FILES_CALIBRATED="${XINTERACT_FILES_CALIBRATED} ${FN_BASE}.recal.indexed.pep.xml"

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet "
$T "${TOOLS_DIR}/comet.2018012.linux.exe" "-D${DATA_DIR}/${FASTA}" "-P${TOOLS_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}.mzML" >/dev/null

echo "Converting to mzID "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Computing recalibration "
$T "${TARGET_DIR}/mzrecal" -scoreFilter="MS:1002257(0.0:0.05)" "${DATA_DIR}/${FN_BASE}.mzML"
echo -n "Creating recalibrated output "
$T "${TARGET_DIR}/mzrecal" -recal "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running msconvert (generating indexed mzml) "
$T "${TOOLS_DIR}/msconvert" "${DATA_DIR}/${FN_BASE}.recal.mzML" --outfile "${FN_BASE}.recal.indexed.mzML" -o "${DATA_DIR}"  >/dev/null
echo -n "Running comet on recalibrated output "
$T "${TOOLS_DIR}/comet.2018012.linux.exe" -D${DATA_DIR}/${FASTA} "-P${TOOLS_DIR}/comet.params" "${DATA_DIR}/${FN_BASE}.recal.indexed.mzML"  >/dev/null

echo "Converting to mzID "
"$TOOLS_DIR/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.recal.indexed.pep.xml" >/dev/null 2>/dev/null
# Note: idconvert does not handle files with multiple extensions correctly,
# hence the .indexed extension is missing
echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.recal.mzid" | wc -l

done

THISDIR=$PWD
cd "${DATA_DIR}"
echo -n "Running peptide prophet (xinteract) on uncalibrated files ${XINTERACT_FILES_UNCALIBRATED}"
$T "${TOOLS_DIR}/xinteract" -Ninteract.pep.xml -p0.05 -l7 -PPM -O ${XINTERACT_FILES_UNCALIBRATED}
echo -n "Running peptide prophet (xinteract) on calibrated files ${XINTERACT_FILES_CALIBRATED}"
$T "${TOOLS_DIR}/xinteract" -Ninteract.recal.pep.xml -p0.05 -l7 -PPM -O ${XINTERACT_FILES_CALIBRATED}
cd "${THISDIR}"
