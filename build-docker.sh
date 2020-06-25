#!/bin/bash
TARGET_DIR=~/mzrecal

mkdir -p $TARGET_DIR

# Create Docker image
cp Dockerfile $TARGET_DIR
( DIR=$PWD; cd $TARGET_DIR ; docker build --tag mzrecal . )

# Use as:
# docker run -v /home/robm/data:/data  mzrecal /mzrecal  -scorefilter="MS:1002257(0.0:1e0)" -mzid="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" -cal="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.recal.json" "/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML"
