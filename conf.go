package main

import "flag"

type fileType int

const (
	Quaver fileType = iota
	Osu
)

type ProgramConfig struct {
	Verbose           bool
	Input             string
	Output            string
	Volume            int
	FileType          fileType
	HPDrain           float64
	OverallDifficulty float64
	KeepSubtitles     bool
	NoStoryboard      bool
	NoMeasureLines    bool
	JSONOutput        bool
	NoTimingPoints    bool
	NoScratchLane     bool
	JSONOnly          bool
	NoZip             bool
	//SpecialAlignment  bool
}

func NewProgramConfig() *ProgramConfig {
	i := flag.String("i", "-", "Input folder containing BMS FOLDERS, NOT zip files!")
	o := flag.String("o", "-", "Which folder you want the files to be output to")
	vol := flag.Int("vol", 100, "How loud the key sounds should be (0-100 is acceptable, 100 default)")
	fileTypeWanted := flag.String("type", "quaver", "Which file type to use. (quaver | osu)")
	hp := flag.Float64("hp", 8.5, "If file type is 'osu', the HP drain (0-10, 8.5 default)")
	od := flag.Float64("od", 8.0, "If file type is 'osu', the overall difficulty (0-10, 8 default)")
	verbose := flag.Bool("v", false, "If specified, all logs will be shown.")
	keepSubtitles := flag.Bool("keep-subtitles", false, "If this is specified, all implicit subtitles will be removed from the title of the map.")
	noScratch := flag.Bool("auto-scratch", false, " If this is specified, all notes in the scratch lane will be replaced with sound effects instead, and the scratch lane will not be used.")
	noStoryboard := flag.Bool("no-storyboard", false, "If file type is 'osu', and this is specified, background animation elements will be ignored.")
	noMeasureLines := flag.Bool("no-measure-lines", false, "If this is specified, timing points will NOT be added at the end of each track to create visible measure lines. (It's a cosmetic thing and doesn't affect gameplay)")
	noTimingPoints := flag.Bool("no-timing-points", false, "If this is specified then BPM changes will not exist. Helpful for maps whose bpm changes don't load correctly (This is equivalent to no SV)")
	jsonOutput := flag.Bool("json", false, " If this is specified, file data will be output to a json file, which is put into the output folder.")
	jsonOnly := flag.Bool("json-only", false, "When specified, no zips will be created, only .json files.")
	noZip := flag.Bool("no-zip", false, "Skip creating .qp, .osz archive; leave output as folder")

	// TODO: Implement 5K+1 alignment feature someday
	//specialAlignment := flag.String("5k-alignment", "right", "If the style is 5K+1, where should the notes be aligned to? (left for 1-5, right for 3-7. Default is right.)")
	flag.Parse()

	fType := Quaver
	if *fileTypeWanted == "osu" {
		fType = Osu
	}
	return &ProgramConfig{
		Input:             *i,
		Output:            *o,
		Volume:            ClampInt(*vol, 100, 0),
		Verbose:           *verbose,
		FileType:          fType,
		HPDrain:           ClampFloat(*hp, 10.0, 0.0),
		OverallDifficulty: ClampFloat(*od, 10.0, 0.0),
		KeepSubtitles:     *keepSubtitles,
		NoStoryboard:      *noStoryboard,
		NoMeasureLines:    *noMeasureLines,
		JSONOutput:        *jsonOutput,
		NoTimingPoints:    *noTimingPoints,
		NoScratchLane:     *noScratch,
		JSONOnly:          *jsonOnly,
		NoZip:             *noZip,
	}
}
