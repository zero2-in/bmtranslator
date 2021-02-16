package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

func (conf *ProgramConfig) ConvertBmsToQua(fileData FileData, outputPath string) error {
	quaFile, e := os.Create(outputPath)
	if e != nil {
		return e
	}
	defer quaFile.Close()

	// flush contents to qua
	_ = WriteLine(quaFile, "AudioFile: virtual")
	_ = WriteLine(quaFile, "SongPreviewTime: -1")
	bg := fileData.Meta.StageFile
	// always prefer banner in the quaver client because of the way song previews are displayed
	if len(fileData.Meta.Banner) > 0 {
		bg = fileData.Meta.Banner
	}
	_ = WriteLine(quaFile, "BackgroundFile: "+bg)
	_ = WriteLine(quaFile, "MapId: -1")
	_ = WriteLine(quaFile, "MapSetId: -1")
	_ = WriteLine(quaFile, "Mode: Keys7")
	scratchKey := "True"
	if conf.NoScratchLane {
		scratchKey = "False"
	}
	_ = WriteLine(quaFile, fmt.Sprintf("HasScratchKey:%s", scratchKey))
	_ = WriteLine(quaFile, fmt.Sprintf("Title: '%s'", fileData.Meta.Title))
	_ = WriteLine(quaFile, fmt.Sprintf("Artist: '%s'", fileData.Meta.Artist))
	_ = WriteLine(quaFile, "Source: BMS")
	_ = WriteLine(quaFile, fmt.Sprintf("Tags: '%s'", fileData.Meta.Tags))
	_ = WriteLine(quaFile, fmt.Sprintf("Creator: '%s'", AppendSubartistsToArtist(fileData.Meta.Artist, fileData.Meta.Subartists)))
	_ = WriteLine(quaFile, fmt.Sprintf("DifficultyName: '%s'", GetDifficultyName(fileData.Meta.Difficulty, fileData.Meta.Subtitle, conf.NoScratchLane)))
	_ = WriteLine(quaFile, "Description: Converted from BMS")
	_ = WriteLine(quaFile, "EditorLayers: []")
	// Process Hit Sound Paths
	_ = WriteLine(quaFile, "CustomAudioSamples:")
	for _, m := range fileData.SoundStringArray {
		_ = WriteLine(quaFile, "- Path: "+m)
	}
	// Process Sound Effects
	_ = WriteLine(quaFile, "SoundEffects:")
	for _, s := range fileData.SoundEffects {
		_ = WriteLine(quaFile, "- StartTime: "+strconv.Itoa(int(s.StartTime)))
		_ = WriteLine(quaFile, "  Sample: "+strconv.Itoa(s.Sample))
		_ = WriteLine(quaFile, "  Volume: "+strconv.Itoa(s.Volume))
	}
	// Process Timing Points
	_ = WriteLine(quaFile, "TimingPoints:")

	keys := make([]float64, len(fileData.TimingPoints))
	i := 0
	for k := range fileData.TimingPoints {
		keys[i] = k
		i++
	}
	sort.Float64s(keys)
	for _, k := range keys {
		_ = WriteLine(quaFile, fmt.Sprintf("- StartTime: %f", k))
		_ = WriteLine(quaFile, fmt.Sprintf("  Bpm: %f", fileData.TimingPoints[k]))
	}

	// Process Slider Velocities
	_ = WriteLine(quaFile, "SliderVelocities: []")
	// Process Hit Objects
	_ = WriteLine(quaFile, "HitObjects:")
	for lane, objects := range fileData.HitObjects {
		if lane == 8 && conf.NoScratchLane {
			continue
		}
		for _, obj := range objects {
			_ = WriteLine(quaFile, "- StartTime: "+strconv.Itoa(int(obj.StartTime)))
			_ = WriteLine(quaFile, "  Lane: "+strconv.Itoa(lane))
			if obj.IsLongNote && int(obj.EndTime) > int(obj.StartTime) {
				_ = WriteLine(quaFile, "  EndTime: "+strconv.Itoa(int(obj.EndTime)))
			}
			if obj.KeySounds != nil {
				_ = WriteLine(quaFile, "  KeySounds:")
				_ = WriteLine(quaFile, "  - Sample: "+strconv.Itoa(obj.KeySounds.Sample))
				_ = WriteLine(quaFile, "    Volume: "+strconv.Itoa(obj.KeySounds.Volume))
			}
		}
	}

	e = quaFile.Sync()
	if e != nil {
		return e
	}

	return nil
}
