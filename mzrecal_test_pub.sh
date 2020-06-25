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

# Tuning parameters

CALPEAKS=20
# for PPMERR in 3 5 7 10 15 20 30 50
#   do
#     for CALPEAKS in 3 4 5 6 8
#     do


# Avoid locale errors in idconvert
export LC_ALL=C

# Extensions of files that we will create
# To avoid problems, we remove them before starting
RM_EXT='.pep.xml .mzid -recal.mzML -recal.mzid'

FN_BASE=${FN1}
PPMERR=10

# Convert raw data to mzML
# docker run -it --rm -e WINEDEBUG=-all -v ${DATA_DIR}:/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor" /data/${FN1}.wiff

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet "
$T "${COMET}" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet_${PPMERR}ppm.params" "${DATA_DIR}/${FN_BASE}.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Recalibration "
$T "${TOOLS_DIR}/mzrecal" -ppmuncal=${PPMERR} -calmult=${CALPEAKS} "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet on recalibrated output "
$T "${COMET}" -D${DATA_DIR}/${FASTA} "-P${DATA_DIR}/comet_${PPMERR}ppm.params" "${DATA_DIR}/${FN_BASE}-recal.mzML"

echo "Converting to .mzid  "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l


FN_BASE=${FN2}
PPMERR=5

# Convert raw data to mzML
# docker run -it --rm -e WINEDEBUG=-all -v ${DATA_DIR}:/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor" /data/${FN1}.raw

echo "Removing intermediate files from previous runs"
for E in ${RM_EXT}; do rm -f "${DATA_DIR}/${FN_BASE}${E}"; done

echo "Input file ${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet "
$T "${COMET}" "-D${DATA_DIR}/${FASTA}" "-P${DATA_DIR}/comet_${PPMERR}ppm.params" "${DATA_DIR}/${FN_BASE}.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}.mzid" | wc -l

echo -n "Recalibration "
$T "${TOOLS_DIR}/mzrecal" -ppmuncal=${PPMERR} -calmult=${CALPEAKS} "${DATA_DIR}/${FN_BASE}.mzML"

echo -n "Running comet on recalibrated output "
$T "${COMET}" -D${DATA_DIR}/${FASTA} "-P${DATA_DIR}/comet_${PPMERR}ppm.params" "${DATA_DIR}/${FN_BASE}-recal.mzML"

echo "Converting to .mzid "
"${TOOLS_DIR}/idconvert" -o "${DATA_DIR}" "${DATA_DIR}/${FN_BASE}-recal.pep.xml" >/dev/null 2>/dev/null

echo -n "Number of identifications with expectation value<0.01: "
grep 'Comet:expectation value" value=".*E-.[3-9]' -P "${DATA_DIR}/${FN_BASE}-recal.mzid" | wc -l


FN_BASE=${FN1}
PPMERR=10
"${TOOLS_DIR}/plot-recal.R" -e "${EXPECT}"  -m ${PPMERR} --outfile="${DATA_DIR}/${FN_BASE}" \
   "--name=TOF data (${FN_BASE})" "${DATA_DIR}/${FN_BASE}.mzid" "${DATA_DIR}/${FN_BASE}-recal.mzid"

FN_BASE=${FN2}
PPMERR=5
"${TOOLS_DIR}/plot-recal.R" -e "${EXPECT}" --nolegend -m ${PPMERR} --outfile="${DATA_DIR}/${FN_BASE}" \
   "--name=Orbitrap data (${FN_BASE})" "${DATA_DIR}/${FN_BASE}.mzid" "${DATA_DIR}/${FN_BASE}-recal.mzid"

montage "${DATA_DIR}/${FN2}.png" "${DATA_DIR}/${FN1}.png" -tile 2x1 -geometry +0+0 "${DATA_DIR}/combined.png"

# Convert to 350 dpi (or 1200 for line art) .jpg, .gif, .tif or .eps
# -> jpg: will not be accepted according to Oxford guidelines
# -> gif, tif: does not work in latexpdf converter,
#              including size (natwidth=2452,natheight=1226) gets a bit further
#              but still does not work
# -> eps: too many elements to use vector graphics
# -> Use png, works with latexpdf only
# All figures should be formatted to fit into, or be
# reduced to, a single (86 mm) or double (178 mm) column width.
# 178/25.4*350=2452.755905512
# -undercolor '#FFFFFFFF' 
convert "${DATA_DIR}/combined.png" -resize 2452\> -fill black  -pointsize 36 \
-annotate +120+86 'A' -annotate +460+86 'B' -annotate +1342+86 'C' -annotate +1685+86 'D' \
"${HOME}/Documents/mzRecal/images/combined.png"

# done
# done
