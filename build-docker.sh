#!/bin/bash
TARGET_DIR=~/mzrecal-docker

mkdir -p $TARGET_DIR

# Obtain version number from git
VERSION=$(git describe --abbrev --dirty --always --tags)
VERSION=${VERSION#"v"}
# build flags are set to creat a binairy reproducable, fully statically linked executable that includes the Git version number
# -trimpath: don't inlude full source code path names in the executable. Needed to produce binairy reproducable output.
# -ldflags:
#   -extldflags \"-static\" : statically link cgo code. This flag is not needed in the current verion and will be removed.
#   -buildid= : clear the buildid, needed to produce binairy reproducable output
#   -X main.progVersion=${VERSION} : include Git version info (from enviroment variable ${VERSION}).
#       This will probably be replaced by using function debug.ReadBuildInfo() from package "runtime/debug" in the future,
#       so that we don't need this build script anymore

echo 'Building mzrecal for Linux/amd64'
GOOS=linux GOARCH=amd64 go build -trimpath -a -ldflags "-extldflags \"-static\" -buildid= -X main.progVersion=${VERSION}" -o $TARGET_DIR/mzrecal

# Create Docker image
cp Dockerfile $TARGET_DIR
( DIR=$PWD; cd $TARGET_DIR ; docker build --tag robmarissen/mzrecal:${VERSION} . )

# Use as:
# docker run -v /home/robm/data:/data  mzrecal /mzrecal  -scorefilter="MS:1002257(0.0:1e0)" -mzid="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" -cal="/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.recal.json" "/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML"
