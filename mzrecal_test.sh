#!/bin/bash
#go run mzrecal.go gsl_wrapper.go  --mzml="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzML" --mzid="/home/robm/data/msrecal_usefull/QE_Plus_YangJing_TCP_iTRAQ_30um_10ul_20170427_R1.mzid"

go run mzrecal.go gsl_wrapper.go --scoreFilter="MS:1002257(0.0:1e0)" --mzml="/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzML" --mzid="/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.mzid" --cal="/home/robm/data/msrecal_ribosomes/human_ribosome_60S_bottomup_peak.recal.json"
