# This makefile is used to run mzrecal on a set of files
#
# Before using:
# - The following tools must be installed (in ${HOME}/tools, modify location below if needed):
#   mzrecal, comet, idconvert, plot-recal.R
# - Set PPM1 to the mass error to search for PSM used for recalibration
# - Set PPM2 to the mass error for comparing the recalibrated results to the original
# - A single comet parameter file must be in the data directory.
#   The 'ppm' parameter is not used, the file is copied and the parameter is set in the copy.
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
KEEP_PEPXML = yes
PARAM_FILE ?= $(wildcard $(DATA_DIR)/*.params)
FASTA ?= $(wildcard $(DATA_DIR)/*.fasta)
TOOLS_DIR=$(HOME)/tools
SEARCHENGINE=$(TOOLS_DIR)/comet.2019015.linux.exe
IDCONVERT=$(TOOLS_DIR)/idconvert
MSCONVERT=$(TOOLS_DIR)/msconvert
MSCONVERT_DOCKER=docker run -it --rm -e WINEDEBUG=-all -v $(DATA_DIR):/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor"
MZRECAL=$(TOOLS_DIR)/mzrecal
MZRECAL_FLAGS=-ppmuncal=$(PPM1)
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
ifdef KEEP_PEPXML
PEPXMLS1 = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-$(PPM1)ppm.pep.xml))
PEPXMLS2 = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-$(PPM2)ppm.pep.xml))
RECALPEPXMLS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.pep.xml))
endif
MZMLSLN = $(addprefix $(RESULT_DIR)/,$(MZMLS))
PLOTS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.png))
TXTSCORE = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.txt))

INTERMEDIATES = $(MZMLSLN) $(PEPS1) $(PEPS2) $(RECALPEPS) $(RECALS) \
 $(PEPXMLS1) $(PEPXMLS2) $(RECALPEPXMLS) \
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
	$(IDCONVERT) -o $(RESULT_DIR)  $(TMP)/$*-$(PPM1)ppm.pep.xml
ifdef KEEP_PEPXML
	mv $(TMP)/$*-$(PPM1)ppm.pep.xml $(RESULT_DIR)
endif
	rm -rf $(TMP)

# Recalibrate
$(RESULT_DIR)/%-recal.mzML: %.mzML $(RESULT_DIR)/%-$(PPM1)ppm.mzid $(MZRECAL)
	$(eval TMP := $(shell mktemp -d))
	$(MZRECAL) $(MZRECAL_FLAGS) -mzid=$(RESULT_DIR)/$*-$(PPM1)ppm.mzid \
	-o=$(TMP)/$*-recal.mzML -cal=$(RESULT_DIR)/$*-recal.json $<
	$(MSCONVERT) -z $(TMP)/$*-recal.mzML -o $(RESULT_DIR)

# Search recalibrated with small mass window
$(RESULT_DIR)/%-recal.mzid: $(RESULT_DIR)/%-recal.mzML
	$(eval TMP := $(shell mktemp -d))
	ln -sf $< $(TMP)/
	sed -E "s/^peptide_mass_tolerance *=.*/peptide_mass_tolerance = $(PPM2)/" ${PARAM_FILE} > $(TMP)/comet.params 
	$(SEARCHENGINE) -D$(FASTA) -P$(TMP)/comet.params $(TMP)/$*-recal.mzML
	$(IDCONVERT) -o $(RESULT_DIR)  $(TMP)/$*-recal.pep.xml
ifdef KEEP_PEPXML
	mv $(TMP)/$*-recal.pep.xml $(RESULT_DIR)
endif
	rm -rf $(TMP)

# Search uncalibrated with small mass window
$(RESULT_DIR)/%-$(PPM2)ppm.mzid: %.mzML
	$(eval TMP := $(shell mktemp -d))
	ln -sf $(DATA_DIR)/$< $(TMP)/$*-$(PPM2)ppm.mzML
	sed -E "s/^peptide_mass_tolerance *=.*/peptide_mass_tolerance = $(PPM2)/" ${PARAM_FILE} > $(TMP)/comet.params 
	$(SEARCHENGINE) -D$(FASTA) -P$(TMP)/comet.params $(TMP)/$*-$(PPM2)ppm.mzML
	$(IDCONVERT) -o $(RESULT_DIR)  $(TMP)/$*-$(PPM2)ppm.pep.xml
ifdef KEEP_PEPXML
	mv $(TMP)/$*-$(PPM2)ppm.pep.xml $(RESULT_DIR)
endif
	rm -rf $(TMP)

# Plot small windows uncalibrated and recalibrated to .png
$(RESULT_DIR)/%.png: $(RESULT_DIR)/%-$(PPM2)ppm.mzid $(RESULT_DIR)/%-recal.mzid 
	$(PLOT) $(PLOT_FLAGS) --name=$* $(RESULT_DIR)/$*-$(PPM2)ppm.mzid $(RESULT_DIR)/$*-recal.mzid \
	--outfile=$(RESULT_DIR)/$* 
