package main

// Layer is the storyboard layer type to use for osu!.
type Layer int

const (
	Back Layer = iota
	Front
)

// BMSFileData shows most things you'd want to know about a bms file.
type BMSFileData struct {
	// Metadata contains the BMS file's metadata.
	Metadata BMSMetadata `json:"metadata"`

	// StartingBPM is what BPM the first track will start with unless it is changed.
	StartingBPM float64

	// LNObject is the designated LN object for this file, unless it is LNTYPE 2,
	// where it will not be used.
	LNObject string

	// TrackLines contains a map of track numbers and their lines.
	TrackLines map[int][]Line

	// HitObjects contains a map of lane # as the key, and hit objects as the value.
	HitObjects map[int][]HitObject

	// TimingPoints contains a map of starting times (in ms) for timing points, and what BPM to
	// change to at that time.
	TimingPoints map[float64]float64

	// SoundEffects contains an array of sound effects to use in the chart.
	SoundEffects []SoundEffect

	// BGAFrames contains an array of background animation frames to use.
	// This is only applicable to osu! or if the file is being output to JSON.
	BGAFrames []BGAFrame

	// Audio contains information about the file's audio. Ref AudioData for more information.
	Audio AudioData

	// Indices contains a list of indexes mapping hexadecimal codes to values.
	Indices IndexData
}

// AudioData contains data about the BMS file's audio, EXCEPT for sound effects, which
// are included at the highest level. This simply contains two lists which map out a hexadecimal
// code to an audio file. Both are separated due to each having a separate purpose.
type AudioData struct {
	StringArray      []string
	HexadecimalArray []string
}

// IndexData contains indices which map hexadecimal codes to values.
type IndexData struct {
	// BPMChanges maps hexadecimal codes to new BPM values.
	BPMChanges map[string]float64

	// Stops maps hexadecimal codes to STOP values.
	Stops map[string]float64

	// BGA maps hexadecimal codes to a file path.
	BGA map[string]string
}

// BGAFrame is a specific BGA frame of the chart.
type BGAFrame struct {
	// StartTime is the precise time, in milliseconds, when this BGA frame should appear.
	StartTime float64 `json:"start_time"`

	// File is the location of the background animation frame as an image.
	File string `json:"file"`

	// Layer is what hypothetical z-index should be used (only applicable to osu!).
	Layer Layer `json:"layer"`
}

// BMSMetadata contains the general metadata of the map, and does not contain any technical
// information.
type BMSMetadata struct {
	// Title is the name of the song.
	Title string `json:"title"`

	// Artist is the composer of the song.
	Artist string `json:"artist"`

	// Tags are used to help find the map wherever it is posted.
	Tags string `json:"tags"`

	// Difficulty is the name of the chart. (This is synonymous to osu!'s "Difficulty Name"
	// metadata field in the editor)
	Difficulty string `json:"difficulty"`

	// StageFile is the main background for the chart, presumably used in BMS clients on
	// song select screen.
	StageFile string `json:"stage_file"`

	// Subtitle is any additional comments left by the map creator and/or artist.
	Subtitle string `json:"subtitle"`

	// SubArtists is an array of usernames who have assisted with map creation.
	// Typically, object/playtester/etc. people are put here.
	SubArtists []string `json:"subartists"`

	// Banner is used in the song select screen and, for some clients, the image
	// which appears while the chart is loading.
	Banner string `json:"banner"`
}

// SoundEffect is a sound effect which will always play at the start time given.
// The player does not have to hit a note for this to trigger.
type SoundEffect struct {
	// StartTime is the time, in milliseconds, where the sound effect should play.
	StartTime float64 `json:"start_time"`

	// Sample is the index of the sample which should be played (see AudioData).
	Sample int `json:"sample"`

	// Volume is the volume of this sound effect. Can be from 0 to 100.
	Volume int `json:"volume"`
}

// HitObject is a note or long note in the chart.
type HitObject struct {
	// StartTime defines the exact millisecond timestamp where the note should be hit at.
	StartTime float64 `json:"start_time"`

	// EndTime is only used if IsLongNote is true, and denotes when the long note ends.
	EndTime float64 `json:"end_time"`

	// IsLongNote is true if the hit object type is a long note (hold and release).
	IsLongNote bool `json:"is_long_note"`

	// KeySounds contains data about the key sound which plays when the player hits this note.
	KeySounds *KeySound `json:"key_sounds,omitempty"`
}

// KeySound represents a sound effect which should be played when the player
// hits a note.
type KeySound struct {
	// Sample is the index of the sample which should be played (see BMSFileData.StringArray).
	Sample int `json:"sample"`

	// Volume is the volume of this key sound. Can be from 0 to 100.
	Volume int `json:"volume"`
}

// Line is a #, followed by channel #, followed by the message.
// A Line is associated with a specific track.
// This is not a definition for headers.
type Line struct {
	// What channel this line represents.
	Channel string `json:"channel"`

	// The message of the line.
	Message string `json:"message"`
}

// LocalBPMChange represents a BPM change (or exBPM) which occurs within a specific track.
// See both https://hitkey.nekokan.dyndns.info/cmds.htm#BPMXX and
// https://hitkey.nekokan.dyndns.info/cmds.htm#EXBPMXX for more information on this.
type LocalBPMChange struct {
	// The precise location of where the BPM change occurs.
	Position float64 `json:"position"`

	// The new BPM value.
	Bpm float64 `json:"bpm"`

	// Whether the BPM value is negative or not. If the BPM is 0.0 this will still be false.
	IsNegative bool `json:"is_negative"`
}

// LocalStop represents a #STOP directive which occurs within a specific track.
// See https://hitkey.nekokan.dyndns.info/cmds.htm#STOPXX for more information on this
// directive.
type LocalStop struct {
	// Duration is how long the #STOP should last for.
	Duration float64 `json:"duration"`

	// Position is the precise location of where the #STOP occurs.
	Position float64 `json:"position"`
}

// LocalTrackData holds information about a track's measure scale, BPM changes, and #STOP
// directives. It does not contain information on where notes should be in the track.
type LocalTrackData struct {
	// MeasureScale is the definition of the length of this track.
	// It is based on 4/4 meter.
	MeasureScale float64 `json:"measure_scale"`

	// A record of all BPM changes which occur in this track.
	BPMChanges []LocalBPMChange `json:"bpm_changes"`

	// A record of all #STOP directives which occur in this track.
	Stops []LocalStop `json:"stops"`
}

type ConversionStatus struct {
	// Name of the folder.
	Name string

	// Success indicates how many files within the folder succeeded.
	Success int

	// Fail indicates how many files within the folder failed or did not convert.
	Fail int

	// Skip is true when the folder was skipped altogether.
	Skip bool
}
