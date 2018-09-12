#!/bin/bash
TARGET_DIR=~/mzrecal

mkdir -p $TARGET_DIR
go build -a -ldflags '-extldflags "-static"' -o $TARGET_DIR/mzrecal


#go run mzrecal.go gsl_wrapper.go  --mzml="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzML" --mzid="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzid"
#go run mzrecal.go gsl_wrapper.go --scoreFilter="MS:1002257(0.0:1e0)" --mzid="/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" --cal="/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.recal.json" "/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML"

$TARGET_DIR/mzrecal -scoreFilter="MS:1002257(0.0:1e0)" "/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML"
$TARGET_DIR/mzrecal -recal "/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML"


# Run test from docker container
# docker run -v /home/robm/data:/data  mzrecal /mzrecal  --scoreFilter="MS:1002257(0.0:1e0)" --mzml="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML" --mzid="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" --cal="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.recal.json"
