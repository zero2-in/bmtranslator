package main

type Layer int

const (
	Back Layer = iota
	Front
)

// most things you'd want to know about a bms file.
type FileData struct {
	Meta                BMSMetadata
	TrackLines          map[int][]Line
	SoundStringArray    []string
	SoundHexArray       []string
	StartingBPM         float64
	LnObject            string
	BPMChangeIndex      map[string]float64
	StopIndex           map[string]float64
	BGAIndex            map[string]string
	SoundEffects        []SoundEffect
	HitObjects          map[int][]HitObject
	TimingPoints        map[float64]float64
	BackgroundAnimation []BackgroundAnimation
}

type BackgroundAnimation struct {
	StartTime float64
	File      string
	Layer     Layer
}

type BMSMetadata struct {
	Title      string
	Artist     string
	Tags       string
	Difficulty string
	StageFile  string
	Subtitle   string
	Subartists []string
}

type SoundEffect struct {
	StartTime float64
	Sample    int
	Volume    int
}

type HitObject struct {
	StartTime  float64
	EndTime    float64
	IsLongNote bool
	KeySounds  *KeySound
}

type KeySound struct {
	Sample int
	Volume int
}

type Line struct {
	Channel string
	Message string
}

type LocalBPMChange struct {
	Position   float64
	Bpm        float64
	IsNegative bool
}

type LocalStop struct {
	Duration float64
	Position float64
}

type LocalTrackData struct {
	MeasureScale float64
	BPMChanges   []LocalBPMChange
	Stops        []LocalStop
}
