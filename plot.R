#!/usr/bin/Rscript
# Usage:
#   plot_errors [-f <searchFile1> --file <searchFile2>

suppressPackageStartupMessages(library(futile.logger))
suppressPackageStartupMessages(library(optparse))
suppressPackageStartupMessages(library(MSnbase))
suppressPackageStartupMessages(library(ggplot2))
suppressPackageStartupMessages(library(gridExtra))

# If running in RStudio, supply "debug" command line args
if (Sys.getenv("RSTUDIO") == "1") {
    options=list("ppmerr" = 12, "exp" = 0.01)
    args=c("/home/robm/data/mzrecal-application-note/120118ry_201B7-32_2_2-120118ry007.mzid",
       "/home/robm/data/mzrecal-application-note/120118ry_201B7-32_2_2-120118ry007-recal.mzid")
    opt = list("options" = options, "args" = args)
} else {
    # Parse command line arguments
    optionList <- list(
    make_option(c("-e", "--exp"), default = 0.01, action='store',
                help = "Comet expectation value limit  [default %default]"),
    make_option(c("-m", "--ppmerr"), default = 10, action='store',
                help = "Maximum m/z ppm error to plot [default %default]"),
    make_option(c("-o", "--outfile"), default = "", action='store',
                help = "Output filename [default first input filename]. The file extention is not used.")

    )

    parser <- OptionParser(option_list = optionList)
    opt <- parse_args2(parser)
}

outputFnBase <- ""
# if no outputfile is specified, use the base name of the input file
if (opt$options$outfile == "") {
    outputFnBase <- tools::file_path_sans_ext(opt$args[1])
} else {
    outputFnBase <- tools::file_path_sans_ext(opt$options$outfile)
}

colors<-c("#D00000", "#20F020")
classNames<-c("original","recalibrated")
maxPpmErr = opt$options$ppmerr

mzidGood <- getMzId(opt$args[1], classNames[1], opt$options$exp, maxPpmErr)
scores <- data.frame("Class" = classNames[[1]],
                     "Mean" = mean(mzidGood$ppmErr),
                     "SD" = sd(mzidGood$ppmErr),
                     "Count" = length(mzidGood$ppmErr))

for (i in c(2:length(opt$args))) {
    mzidGoodX <- getMzId(opt$args[i], classNames[i], opt$options$exp, maxPpmErr)
    scoresX <- data.frame("Class" = classNames[[i]],
                         "Mean" = mean(mzidGoodX$ppmErr),
                         "SD" = sd(mzidGoodX$ppmErr),
                         "Count" = length(mzidGoodX$ppmErr))
    scores <- rbind(scores, scoresX)
    
    mzidGood <- rbind(mzidGood, mzidGoodX)

}

scoreFile <- paste(outputFnBase, ".txt", sep="")
sink(scoreFile)
perfScore <- (scores$Mean[1]/scores$Mean[2] +
          3*scores$SD[1]/scores$SD[2] +
          10*scores$Count[2]/scores$Count[1])
print(perfScore[[1]])
print(scores)
sink()

# Shuffle rows, so that overlapping points get approximatly fair color in plot
set.seed(42)
rows <- sample(nrow(mzidGood))
mzidGood <- mzidGood[rows, ]

g <- ggplot(mzidGood, aes(x=calculatedMassToCharge, y=ppmErr, colour = class))+
        theme(axis.title.y=element_blank(),
        axis.text.y=element_blank(),
        axis.ticks.y=element_blank()) +
        geom_point(size=1, alpha = 0.3) +
        scale_color_manual(values = colors)

gd = ggplot(mzidGood, aes(x=ppmErr, colour = class))  +
        geom_density() +
        coord_flip() +
        theme(legend.position = "none") +
        scale_color_manual(values = colors)

#grid.arrange(gd, g, widths = c(1, 2), ncol=2)
p <- arrangeGrob(gd, g, widths = c(1, 2), ncol=2)

plotFile <- paste(outputFnBase, ".svg", sep="")

ggsave(
  plotFile,
  plot = p,
  device = NULL,
  path = NULL,
  scale = 1,
  width = NA,
  height = NA,
  units = c("in", "cm", "mm"),
  dpi = 300,
  limitsize = TRUE,
)

getMzId<-function(fileName, className, cometExpLim, maxPpmErr)
{
    mzid <- readMzIdData(fileName)
    mzidGood=subset(mzid, Comet.expectation.value < cometExpLim)
    mzidGood$mzErr <- mzidGood$calculatedMassToCharge - mzidGood$experimentalMassToCharge

    # FIXME: Handle the case where precursor is of by one (abs(mzError)= 0.5 or 0.3333 or ...)
    mzidGoodX=subset(mzidGood, (1000000.0*abs(mzErr)/calculatedMassToCharge) < maxPpmErr)
    mzidGoodX$ppmErr <- 1000000.0* mzidGoodX$mzErr / mzidGoodX$calculatedMassToCharge
    mzidGoodX$class=className
    mzidGoodX
}



