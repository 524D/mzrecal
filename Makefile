# This makefile is used to run mzrecal on a set of files
#
# Before using:
# - The following tools must be installed (in ${HOME}/tools, modify below if needed):
#   mzrecal, comet, idconvert, plot-recal.R
# - Set PPM1 to the mass error to search for PSM used for recalibration
# - Set PPM2 to the mass error for comparing the recalibrated results to the original
# - Comet parameter files must be in same directory as search data, and must be
#   named: <dontcare>${PPM1}ppm.params and <dontcare>${PPM1}ppm.params
# 


# It performs the following steps:
# Identify peptides (using comet)
# Convert .pep.XML into .mzid (using idconvert)
# Recalibrate (using mzrecal)
# Add index to mzML file (using msconvert)
# Identify peptides in recalibrated files (using comet)
# Combine results from multiple files, both before and after recalibration
#  (using peptide prophet/xinteract)
# Display a histogram of ms2 precursor mass errors before and
#  after recalibration
# All results and intermediate results are stored in a separate direcory
# defined by RESULT_DIR
DATA_DIR=$(shell pwd)
DATA_BASE=$(shell basename $(DATA_DIR))
RESULT_DIR=$(HOME)/results/$(DATA_BASE)
# m/z error to find PSM's for recalibration
PPM1=10
# m/z error to compare recalibrated with original
PPM2=5
PARAM_FILE ?= $(wildcard $(DATA_DIR)/*.params)
FASTA ?= $(wildcard $(DATA_DIR)/*.fasta)
TOOLS_DIR=$(HOME)/tools
SEARCHENGINE=$(TOOLS_DIR)/comet.2019015.linux.exe
IDCONVERT=$(TOOLS_DIR)/idconvert
MSCONVERT_DOCKER=docker run -it --rm -e WINEDEBUG=-all -v $(DATA_DIR):/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor"
MZRECAL=$(TOOLS_DIR)/mzrecal
MZRECAL_FLAGS=-mzTry=$(PPM1)
PLOT=$(TOOLS_DIR)/plot-recal.R
PLOT_FLAGS=-e 0.01 -m $(PPM2)

# Avoid locale errors in idconvert
export LC_ALL=C

MZMLS = $(wildcard *.mzML)
MZIDS = $(MZMLS:.mzML=.mzid)
RECALS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.mzML))
PEPS1 = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-$(PPM1)ppm.mzid))
PEPS2 = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-$(PPM2)ppm.mzid))
RECALPEPS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.mzid))
MZMLSLN = $(addprefix $(RESULT_DIR)/,$(MZMLS))
PLOTS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.png))
TXTSCORE = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.txt))

INTERMEDIATES = $(MZMLSLN) $(PEPS1) $(PEPS2) $(RECALPEPS) $(RECALS) \
 $(RECALS:-recal.mzML=-recal.json)

# Main target
all: dirs $(RECALS) $(PLOTS)

# To remove generated files
clean:
	rm -f $(RECALS) $(PLOTS) $(TXTSCORE) $(INTERMEDIATES)

# Prevent deletion of intermediate files
.SECONDARY: $(INTERMEDIATES)

.PHONY: dirs

dirs: $(RESULT_DIR)

$(RESULT_DIR):
	mkdir -p $(RESULT_DIR)

$(RECALPEPS): $(RECALS)

# Search uncalibrated with wide mass window
# Since 'comet' always names its output file after the input file,
# and 'idconvert' always names its output file after the mzML file,
# we put intermediate files in a temproary directory to get this
# working. 
$(RESULT_DIR)/%-$(PPM1)ppm.mzid: %.mzML
	$(eval TMP := $(shell mktemp -d))
	ln -sf $(DATA_DIR)/$< $(TMP)/$*-$(PPM1)ppm.mzML
	# Create updated comet parameter file with desired PPM error
	sed -E "s/^peptide_mass_tolerance *=.*/peptide_mass_tolerance = $(PPM1)/" ${PARAM_FILE} > $(TMP)/comet.params 
	$(SEARCHENGINE) -D$(FASTA) -P$(TMP)/comet.params $(TMP)/$*-$(PPM1)ppm.mzML
	$(IDCONVERT) -o $(TMP)  $(TMP)/$*-$(PPM1)ppm.pep.xml
	mv $(TMP)/*-$(PPM1)ppm.mzid $(RESULT_DIR)
	rm -rf $(TMP)

# Recalibrate
$(RESULT_DIR)/%-recal.mzML: %.mzML $(RESULT_DIR)/%-$(PPM1)ppm.mzid $(MZRECAL)
	$(MZRECAL) $(MZRECAL_FLAGS) -mzid=$(RESULT_DIR)/$*-$(PPM1)ppm.mzid \
	-mzmlOut=$@ -cal=$(RESULT_DIR)/$*-recal.json $<

# Search recalibrated with small mass window
$(RESULT_DIR)/%-recal.mzid: $(RESULT_DIR)/%-recal.mzML
	$(eval TMP := $(shell mktemp -d))
	ln -sf $< $(TMP)/
	sed -E "s/^peptide_mass_tolerance *=.*/peptide_mass_tolerance = $(PPM2)/" ${PARAM_FILE} > $(TMP)/comet.params 
	$(SEARCHENGINE) -D$(FASTA) -P$(TMP)/comet.params $(TMP)/$*-recal.mzML
	$(IDCONVERT) -o $(TMP)  $(TMP)/$*-recal.pep.xml
	mv $(TMP)/*-recal.mzid $(RESULT_DIR)
	rm -rf $(TMP)

# Search uncalibrated with small mass window
$(RESULT_DIR)/%-$(PPM2)ppm.mzid: %.mzML
	$(eval TMP := $(shell mktemp -d))
	ln -sf $(DATA_DIR)/$< $(TMP)/$*-$(PPM2)ppm.mzML
	sed -E "s/^peptide_mass_tolerance *=.*/peptide_mass_tolerance = $(PPM2)/" ${PARAM_FILE} > $(TMP)/comet.params 
	$(SEARCHENGINE) -D$(FASTA) -P$(TMP)/comet.params $(TMP)/$*-$(PPM2)ppm.mzML
	$(IDCONVERT) -o $(TMP)  $(TMP)/$*-$(PPM2)ppm.pep.xml
	mv $(TMP)/*-$(PPM2)ppm.mzid $(RESULT_DIR)
	rm -rf $(TMP)

# Plot small windows uncalibrated and recalibrated to .png
$(RESULT_DIR)/%.png: $(RESULT_DIR)/%-$(PPM2)ppm.mzid $(RESULT_DIR)/%-recal.mzid 
	$(PLOT) $(PLOT_FLAGS) $(RESULT_DIR)/$*-$(PPM2)ppm.mzid $(RESULT_DIR)/$*-recal.mzid
