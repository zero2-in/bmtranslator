package main

type ConvertedFile struct {
	Metadata            BMSMetadata
	SoundEffects        []SoundEffect
	HitObjects          []HitObject
	TimingPoints        map[float64]float64
	KeySoundStringArray []string
	BackgroundAnimation []BackgroundAnimation
}

type Layer int

const (
	Back Layer = iota
	Front
)

type BackgroundAnimation struct {
	StartTime float64
	File      string
	Layer     Layer
}

type BMSMetadata struct {
	Title      string
	Artist     string
	Tags       string
	Creator    string
	Difficulty string
	StageFile  string
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
	Lane       int
	KeySounds  *HitObjectKeySound
}

type HitObjectKeySound struct {
	Sample int
	Volume int
}

type LocalTrackData struct {
	Channel string
	Message string
}

type LocalTempoChange struct {
	Position   float64
	Bpm        float64
	IsNegative bool
}

type LocalStopCommand struct {
	Duration float64
	Position float64
}
