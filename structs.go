package main

type Layer int

const (
	Back Layer = iota
	Front
)

// FileData shows most things you'd want to know about a bms file.
type FileData struct {
	Metadata            BMSMetadata `json:"metadata"`
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
	StartTime float64 `json:"start_time"`
	File      string  `json:"file"`
	Layer     Layer   `json:"layer"`
}

type BMSMetadata struct {
	Title      string   `json:"title"`
	Artist     string   `json:"artist"`
	Tags       string   `json:"tags"`
	Difficulty string   `json:"difficulty"`
	StageFile  string   `json:"stage_file"`
	Subtitle   string   `json:"subtitle"`
	Subartists []string `json:"subartists"`
	Banner     string   `json:"banner"`
}

type SoundEffect struct {
	StartTime float64 `json:"start_time"`
	Sample    int     `json:"sample"`
	Volume    int     `json:"volume"`
}

type HitObject struct {
	StartTime  float64   `json:"start_time"`
	EndTime    float64   `json:"end_time"`
	IsLongNote bool      `json:"is_long_note"`
	KeySounds  *KeySound `json:"key_sounds,omitempty"`
}

type KeySound struct {
	Sample int `json:"sample"`
	Volume int `json:"volume"`
}

type Line struct {
	Channel string `json:"channel"`
	Message string `json:"message"`
}

type LocalBPMChange struct {
	Position   float64 `json:"position"`
	Bpm        float64 `json:"bpm"`
	IsNegative bool    `json:"is_negative"`
}

type LocalStop struct {
	Duration float64 `json:"duration"`
	Position float64 `json:"position"`
}

type LocalTrackData struct {
	MeasureScale float64          `json:"measure_scale"`
	BPMChanges   []LocalBPMChange `json:"bpm_changes"`
	Stops        []LocalStop      `json:"stops"`
}

type ConversionStatus struct {
	Name    string
	Success int
	Fail    int
	Skip    bool
}
