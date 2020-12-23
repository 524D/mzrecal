#!/usr/bin/Rscript
# Usage:
#   plot_errors [-f <searchFile1> --file <searchFile2>

suppressPackageStartupMessages(library(stringr))
suppressPackageStartupMessages(library(futile.logger))
suppressPackageStartupMessages(library(optparse))
suppressPackageStartupMessages(library(MSnbase))
suppressPackageStartupMessages(library(ggplot2))
suppressPackageStartupMessages(library(grid))
suppressPackageStartupMessages(library(gridExtra))


getMzId<-function(fileName, className, cometExpLim, maxPpmErr)
{
    mzid <- readMzIdData(fileName)
    mzidGood <- subset(mzid, Comet.expectation.value < cometExpLim)

    psms <- nrow(mzidGood)


    # Only keep columns that we want
    keeps <- c( "experimentalMassToCharge", "calculatedMassToCharge")
    mzidGood=mzidGood[keeps]

    mzidGood$mzErr <- mzidGood$experimentalMassToCharge - mzidGood$calculatedMassToCharge
    mzidGoodX=subset(mzidGood, (1000000.0*abs(mzErr)/calculatedMassToCharge) < maxPpmErr)
    mzidGoodX$ppmErr <- 1000000.0* mzidGoodX$mzErr / mzidGoodX$calculatedMassToCharge
    mzidGoodX$class=className

    mzIdVals <- list("psms" = psms, "mzidGoodX" = mzidGoodX)
    mzIdVals
}

####################################################################################

# If running in RStudio, supply "debug" command line args
if (Sys.getenv("RSTUDIO") == "1") {
    options=list("ppmerr" = 10, "exp" = 0.01, "nolegend" = FALSE, "outfile" = "")

    args=c("/home/robm/results/PXD000153/20011221_04_BarH_IM2_10to11.mzid",
       "/home/robm/results/PXD000153/20011221_04_BarH_IM2_10to11-recal.mzid")
    opt = list("options" = options, "args" = args)
} else {
    # Parse command line arguments
    optionList <- list(
    make_option(c("-e", "--exp"), default = 0.01, action='store',
                help = "Comet expectation value limit  [default %default]"),
    make_option(c("-m", "--ppmerr"), default = 10, action='store',
                help = "Maximum m/z ppm error to plot [default %default]"),
    make_option(c("-o", "--outfile"), default = "", action='store',
                help = "Output filename [default first input filename]. The file extention is not used."),
    make_option(c("-L", "--nolegend"), default = FALSE, action='store_true',
                help = "If set, no legend is shown."),
    make_option(c("-n", "--name"), default = "", action='store',
                help = "Name of data, printed at top of picture.")

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

# Test colors with: https://www.color-blindness.com/coblis-color-blindness-simulator/
# Use magenta instead of red, that's better according to some sources
colors<-c("#F000A0", "#00B000")
classNames<-c("original","recalibrated")
maxPpmErr = opt$options$ppmerr

mzIdVals <- getMzId(opt$args[1], classNames[1], opt$options$exp, maxPpmErr)
mzidGood <- mzIdVals$mzidGood
psms <- mzIdVals$psms
# Keep track of density of error distribution, we need it later
dens <- list()
dens[[length(dens)+1]] <- density(mzidGood$ppmErr)
scores <- data.frame("Class" = classNames[[1]],
                     "Mean" = mean(mzidGood$ppmErr),
                     "SD" = sd(mzidGood$ppmErr),
                     "Count" = length(mzidGood$ppmErr))

for (i in c(2:length(opt$args))) {
    mzIdVals <- getMzId(opt$args[i], classNames[i], opt$options$exp, maxPpmErr)
    psms <- c(psms, mzIdVals$psms)
    dens[[length(dens)+1]] <- density(mzIdVals$mzidGood$ppmErr)
    # Append the good PSM's (with different class name)
    mzidGood <- rbind(mzidGood, mzIdVals$mzidGood)
    scoresX <- data.frame("Class" = classNames[[i]],
                         "Mean" = mean( mzIdVals$mzidGood$ppmErr),
                         "SD" = sd( mzIdVals$mzidGood$ppmErr),
                         "Count" = length( mzIdVals$mzidGood$ppmErr))
    scores <- rbind(scores, scoresX)
}

# Create a text file with some numbers that indicate how well the recalibration worked
scoreFile <- paste(outputFnBase, ".txt", sep="")
sink(scoreFile)
perfScore <- (scores$Mean[1]/scores$Mean[2] +
          3*scores$SD[1]/scores$SD[2] +
          10*scores$Count[2]/scores$Count[1])
print(perfScore[[1]])
print(scores)
sink()

# Before plotting, shuffle rows so that overlapping points get approximately fair color in plot
set.seed(42)
rows <- sample(nrow(mzidGood))
mzidGood <- mzidGood[rows, ]

# Make labels with number of PSMs in each class
txt1 = paste("n=", psms[1], sep="")
txt2 = paste("n=", psms[2], sep="")
massScaleTxt <- "mass error (ppm)";
# Position of labels with PSMs
x1<- dens[[1]]$x[which.max(dens[[1]]$y)]
y1_top = dens[[1]]$y[which.max(dens[[1]]$y)]
y1<- y1_top * 0.4;
x2<- dens[[2]]$x[which.max(dens[[2]]$y)]
y2_top = dens[[2]]$y[which.max(dens[[2]]$y)]
y2<- y2_top * 0.4;

# Ensure labels are separated vertically
if (abs(x1-x2)<0.05 * maxPpmErr) {
    if (x1 > 0.05 * maxPpmErr) {
        x1 = x2 - 0.05 * maxPpmErr
    } else {
        x2 = x1 + 0.05 * maxPpmErr
    }
}

# Special case for files used in publication
if (str_detect(outputFnBase, "120118ry_201B7-32_2_2")) {
x1<- -7.1;
y1<- 0.12;
x2<- 2.5;
y2<- 0.12;
massScaleTxt <- "";
}
if (str_detect(outputFnBase, "GSC11_24h_R1")) {
x1<- 1.25;
y1<- 0.45;
x2<- -1.25;
y2<- 0.45;
}

myLegendPos <- "none";
if (!opt$options$nolegend) {
    myLegendPos <- c(0.85, 0.91);
}
g <- ggplot(mzidGood, aes(x=calculatedMassToCharge, y=ppmErr, colour = class))+
        theme(axis.title.y=element_blank(),
        axis.text.y=element_blank(),
        axis.ticks.y=element_blank()) +
        geom_point(size=0.6, alpha = 0.2) +
        theme(text=element_text(size=12, family="sans"),
              legend.title=element_blank(),
              legend.background = element_rect(fill=alpha('white', 0.0)),
              legend.position = myLegendPos) +
        scale_x_continuous(name=expression(italic("m/z")), limits=c(270, 1250)) +
        scale_y_continuous(limits=c(-maxPpmErr, maxPpmErr)) +
        scale_color_manual(values = colors) +
        geom_smooth(method = "lm")


gd = ggplot(mzidGood, aes(x=ppmErr, colour = class))  +
        geom_density() +
        coord_flip() +
        theme(legend.position = "none",
              text=element_text(size=12, family="sans")) +
        scale_x_continuous(name=massScaleTxt, limits=c(-maxPpmErr, maxPpmErr)) +
        scale_color_manual(values = colors) +
        annotate(geom="text", x1, y1, label=txt1, color=colors[1]) +
        annotate(geom="text", x2, y2, label=txt2, color=colors[2])

p <- arrangeGrob(gd, g, widths = c(1, 2), ncol=2,
    top = textGrob(opt$options$name,gp=gpar(fontsize=12)))

plotFile <- paste(outputFnBase, ".png", sep="")


ggsave(
  plotFile,
  plot = p,
  device = NULL,
  path = NULL,
  scale = 1,
  width = 8,
  height = 6,
  units = "in",
  dpi = 300,
  limitsize = TRUE,
)

