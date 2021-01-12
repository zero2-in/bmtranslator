package main

import (
	"fmt"
	"os"
	"sort"
	"strconv"
)

func ConvertBmsToQua(convertedFile ConvertedFile, outputPath string) error {
	quaFile, e := os.Create(outputPath)
	if e != nil {
		return e
	}
	defer quaFile.Close()

	// flush contents to qua
	_ = WriteLine(quaFile, "AudioFile: virtual")
	_ = WriteLine(quaFile, "SongPreviewTime: -1")
	_ = WriteLine(quaFile, "BackgroundFile: "+convertedFile.Metadata.StageFile)
	_ = WriteLine(quaFile, "MapId: -1")
	_ = WriteLine(quaFile, "MapSetId: -1")
	_ = WriteLine(quaFile, "Mode: Keys7")
	_ = WriteLine(quaFile, "HasScratchKey: True")
	_ = WriteLine(quaFile, "Title: "+convertedFile.Metadata.Title)
	_ = WriteLine(quaFile, "Artist: "+convertedFile.Metadata.Artist)
	_ = WriteLine(quaFile, "Source: BMS")
	_ = WriteLine(quaFile, "Tags: "+convertedFile.Metadata.Tags)
	_ = WriteLine(quaFile, "Creator: "+convertedFile.Metadata.Creator)
	_ = WriteLine(quaFile, "DifficultyName: "+GetDifficultyName(convertedFile.Metadata.Difficulty))
	_ = WriteLine(quaFile, "Description: Converted from BMS")
	_ = WriteLine(quaFile, "EditorLayers: []")
	// Process Hit Sound Paths
	_ = WriteLine(quaFile, "CustomAudioSamples:")
	for _, m := range convertedFile.KeySoundStringArray {
		_ = WriteLine(quaFile, "- Path: "+m)
	}
	// Process Sound Effects
	_ = WriteLine(quaFile, "SoundEffects:")
	for _, s := range convertedFile.SoundEffects {
		_ = WriteLine(quaFile, "- StartTime: "+strconv.Itoa(int(s.StartTime)))
		_ = WriteLine(quaFile, "  Sample: "+strconv.Itoa(s.Sample))
		_ = WriteLine(quaFile, "  Volume: "+strconv.Itoa(s.Volume))
	}
	// Process Timing Points
	_ = WriteLine(quaFile, "TimingPoints:")

	keys := make([]float64, len(convertedFile.TimingPoints))
	i := 0
	for k := range convertedFile.TimingPoints {
		keys[i] = k
		i++
	}
	sort.Float64s(keys)
	for _, k := range keys {
		_ = WriteLine(quaFile, fmt.Sprintf("- StartTime: %f", k))
		_ = WriteLine(quaFile, fmt.Sprintf("  Bpm: %f", convertedFile.TimingPoints[k]))
	}

	// Process Slider Velocities
	_ = WriteLine(quaFile, "SliderVelocities: []")
	// Process Hit Objects
	_ = WriteLine(quaFile, "HitObjects:")
	for _, h := range convertedFile.HitObjects {
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
