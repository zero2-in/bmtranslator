package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

func ConvertBmsToQua(fileData FileData, outputPath string) error {
	quaFile, e := os.Create(outputPath)
	if e != nil {
		return e
	}
	defer quaFile.Close()

	// flush contents to qua
	_ = WriteLine(quaFile, "AudioFile: virtual")
	_ = WriteLine(quaFile, "SongPreviewTime: -1")
	_ = WriteLine(quaFile, "BackgroundFile: "+fileData.Meta.StageFile)
	_ = WriteLine(quaFile, "MapId: -1")
	_ = WriteLine(quaFile, "MapSetId: -1")
	_ = WriteLine(quaFile, "Mode: Keys7")
	_ = WriteLine(quaFile, "HasScratchKey: True")
	_ = WriteLine(quaFile, fmt.Sprintf("Title: '%s'", fileData.Meta.Title))
	_ = WriteLine(quaFile, fmt.Sprintf("Artist: '%s'", fileData.Meta.Artist))
	_ = WriteLine(quaFile, "Source: BMS")
	_ = WriteLine(quaFile, fmt.Sprintf("Tags: '%s'", fileData.Meta.Tags))
	_ = WriteLine(quaFile, fmt.Sprintf("Creator: '%s'", AppendSubartistsToArtist(fileData.Meta.Artist, fileData.Meta.Subartists)))
	_ = WriteLine(quaFile, fmt.Sprintf("DifficultyName: '%s'", GetDifficultyName(fileData.Meta.Difficulty, fileData.Meta.Subtitle)))
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
	for _, h := range fileData.HitObjects {
		_ = WriteLine(quaFile, "- StartTime: "+strconv.Itoa(int(h.StartTime)))
		_ = WriteLine(quaFile, "  Lane: "+strconv.Itoa(h.Lane))
		if h.IsLongNote {
			if h.EndTime == 0.0 {
				// This long note is invalid/doesn't have an ending time.
				continue
			}
			_ = WriteLine(quaFile, "  EndTime: "+strconv.Itoa(int(h.EndTime)))
		}
		if h.KeySounds != nil {
			_ = WriteLine(quaFile, "  KeySounds:")
			_ = WriteLine(quaFile, "  - Sample: "+strconv.Itoa(h.KeySounds.Sample))
			_ = WriteLine(quaFile, "    Volume: "+strconv.Itoa(h.KeySounds.Volume))
		} else {
			_ = WriteLine(quaFile, "  KeySounds: []")
		}
	}

	e = quaFile.Sync()
	if e != nil {
		return e
	}

	return nil
}
