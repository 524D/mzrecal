DATA_DIR=$(shell pwd)
DATA_BASE=$(shell basename $(DATA_DIR))
RESULT_DIR=$(HOME)/results/$(DATA_BASE)
#FASTA=$(DATA_DIR)/18mix_db_plus_contaminants_20081209.fasta
FASTA ?= $(HOME)/data/uniprothuman_20180620.fasta
TOOLS_DIR=$(HOME)/tools

SEARCHENGINE=$(TOOLS_DIR)/comet.2018012.linux.exe
SEARCHENGINE_FLAGS=-D$(FASTA) -P$(TOOLS_DIR)/comet.params
IDCONVERT=$(TOOLS_DIR)/idconvert
MSCONVERT=$(TOOLS_DIR)/msconvert
MZRECAL=$(TOOLS_DIR)/mzrecal
MZRECAL_FLAGS1=-mzTry=10 -mzAccept=3 -scoreFilter="MS:1002257(0.0:0.05)"
MZRECAL_FLAGS2=-recal -empty-non-calibrated

COMBIRECAL = $(RESULT_DIR)/interact-recal.pep.xml
COMBI = $(RESULT_DIR)/interact.pep.xml
SVGERR = $(RESULT_DIR)/ppmerr.svg
MZMLS = $(wildcard *.mzML)
MZIDS = $(MZMLS:.mzML=.mzid)
RECALS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal.mzML))
PEPS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=.pep.xml))
RECALIDXS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal-indexed.mzML))
RECALPEPS = $(addprefix $(RESULT_DIR)/,$(MZMLS:.mzML=-recal-indexed.pep.xml))

INTERMEDIATES = $(PEPS) $(RECALPEPS) $(RECALS) $(RECALIDXS) $(PEPS:.pep.xml=.mzid) $(RECALPEPS.pep.xml=.mzid)) $(RECALS:-recal.mzML=-recal.json) $(RECALS:-recal.mzML=-recal-indexed.mzML)

# Main target
all: dirs $(RECALS) $(SVGERR)
	inkview $(SVGERR) &

# To remove generated files
clean:
	rm -f $(RECALS) $(RECALPEPS) $(INTERMEDIATES)

# Prevent deletion of intermediate files
.SECONDARY: $(INTERMEDIATES)

.PHONY: dirs

dirs: $(RESULT_DIR)

$(RESULT_DIR):
	mkdir -p $(RESULT_DIR)

$(SVGERR): $(COMBIRECAL) $(COMBI)
	$(TOOLS_DIR)/pepXML -nofixmass -expect=0.05 $(COMBI) $(COMBIRECAL)

$(COMBIRECAL): $(RECALPEPS) $(RECALIDXS)
	$(TOOLS_DIR)/xinteract -N$(RESULT_DIR)/interact-recal.pep.xml -p0.05 -l7 -PPM -O $(RECALPEPS)

$(COMBI): $(PEPS) $(MZMLS)
	$(TOOLS_DIR)/xinteract -N$(RESULT_DIR)/interact.pep.xml -p0.05 -l7 -PPM -O $(PEPS)

# mzML to pep.xml
$(RESULT_DIR)/%.pep.xml: %.mzML
	ln -sf $(DATA_DIR)/$< $(RESULT_DIR)/
	$(SEARCHENGINE) $(SEARCHENGINE_FLAGS) $(RESULT_DIR)/$<
	rm $(RESULT_DIR)/$<
$(RESULT_DIR)/%.pep.xml: $(RESULT_DIR)/%.mzML
	$(SEARCHENGINE) $(SEARCHENGINE_FLAGS) $<

# mzML to -indexed.mzML
%-indexed.mzML: %.mzML
	$(MSCONVERT) $< --outdir $(RESULT_DIR) --outfile $@

# pep.xml to mzid
$(RESULT_DIR)/%.mzid: $(RESULT_DIR)/%.pep.xml
	$(IDCONVERT) -o $(RESULT_DIR)  $<

# Compute recalibrate params
$(RESULT_DIR)/%-recal.json: %.mzML $(RESULT_DIR)/%.mzid $(MZRECAL)
	$(MZRECAL) $(MZRECAL_FLAGS1) -mzid=$(RESULT_DIR)/$*.mzid -cal=$@ $<

# Recalibrate
$(RESULT_DIR)/%-recal.mzML: $(RESULT_DIR)/%-recal.json $(MZRECAL)
	$(MZRECAL) $(MZRECAL_FLAGS2) -cal=$(RESULT_DIR)/$*-recal.json -mzmlOut=$@ $*.mzML
