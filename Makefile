# This makefile is used to test mzrecal on a set of files
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
FASTA ?= $(wildcard $(DATA_DIR)/*.fasta)
TOOLS_DIR=$(HOME)/tools
SEARCHENGINE=$(TOOLS_DIR)/comet.2019015.linux.exe
SEARCHENGINE_FLAGS ?= -D$(FASTA) -P$(wildcard $(DATA_DIR)/*.params)
IDCONVERT=$(TOOLS_DIR)/idconvert
MSCONVERT_DOCKER=docker run -it --rm -e WINEDEBUG=-all -v $(DATA_DIR):/data chambm/pwiz-skyline-i-agree-to-the-vendor-licenses wine msconvert --zlib --filter "peakPicking vendor"
MZRECAL=$(TOOLS_DIR)/mzrecal
MZRECAL_FLAGS=-mzTry=10
PLOT=$(TOOLS_DIR)/plot-recal.R
PLOT_FLAGS=-e 0.01 -m 10

# Avoid locale errors in idconvert
export LC_ALL=C

MZMLS = $(wildcard *.mzML)
MZIDS = $(MZMLS:.mzML=.mzid)
RECALS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.mzML))
PEPS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.pep.xml))
RECALPEPS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.pep.xml))
MZMLSLN = $(addprefix $(RESULT_DIR)/,$(MZMLS))
PLOTS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.png))
TXTSCORE = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.txt))

INTERMEDIATES = $(MZMLSLN) $(PEPS) $(RECALPEPS) $(RECALS) \
 $(PEPS:.pep.xml=.mzid) $(RECALPEPS.pep.xml=.mzid) \
 $(RECALPEPS.pep.xml=-recal.mzid) $(RECALS:-recal.mzML=-recal.json)

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

# pep.xml to mzid
$(RESULT_DIR)/%.mzid: $(RESULT_DIR)/%.pep.xml
	$(IDCONVERT) -o $(RESULT_DIR)  $<

# ./mzid and -recal.mzid to .png
$(RESULT_DIR)/%.png: $(RESULT_DIR)/%.mzid $(RESULT_DIR)/%-recal.mzid 
	$(PLOT) $(PLOT_FLAGS) $(RESULT_DIR)/$*.mzid $(RESULT_DIR)/$*-recal.mzid

$(RECALPEPS): $(RECALS)

# mzML to pep.xml
$(RESULT_DIR)/%.pep.xml: %.mzML
	ln -sf $(DATA_DIR)/$< $(RESULT_DIR)/
	$(SEARCHENGINE) $(SEARCHENGINE_FLAGS) $(RESULT_DIR)/$<

$(RESULT_DIR)/%.pep.xml: $(RESULT_DIR)/%.mzML
	$(SEARCHENGINE) $(SEARCHENGINE_FLAGS) $<

# pep.xml to mzid
$(RESULT_DIR)/%.mzid: $(RESULT_DIR)/%.pep.xml
	$(IDCONVERT) -o $(RESULT_DIR)  $<

# Recalibrate
$(RESULT_DIR)/%-recal.mzML: %.mzML $(RESULT_DIR)/%.mzid $(MZRECAL)
	$(MZRECAL) $(MZRECAL_FLAGS) -mzid=$(RESULT_DIR)/$*.mzid -mzmlOut=$@ $(RESULT_DIR)/$<

