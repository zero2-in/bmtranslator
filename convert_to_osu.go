package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"sort"
)

const (
	OsuYPos = 192
)

// Convert a BMS file to .osu (for the game osu!).
func (conf *ProgramConfig) ConvertBmsToOsu(fileData FileData, outputPath string) error {
	osuFile, e := os.Create(outputPath)
	if e != nil {
		return e
	}
	defer osuFile.Close()

	// flush contents to osu
	_ = WriteLine(osuFile, "osu file format v14\n")

	_ = WriteLine(osuFile, "[General]")
	_ = WriteLine(osuFile, "Mode: 3")
	_ = WriteLine(osuFile, "SampleSet: Soft")
	_ = WriteLine(osuFile, "SpecialStyle: 1")
	_ = WriteLine(osuFile, "Countdown: 0")

	_ = WriteLine(osuFile, "[Editor]")
	_ = WriteLine(osuFile, "DistanceSpacing: 1")
	_ = WriteLine(osuFile, "BeatDivisor: 1")
	_ = WriteLine(osuFile, "GridSize: 1")
	_ = WriteLine(osuFile, "TimelineZoom: 1")

	_ = WriteLine(osuFile, "[Metadata]")
	_ = WriteLine(osuFile, fmt.Sprintf("Title:%s", fileData.Meta.Title))
	_ = WriteLine(osuFile, fmt.Sprintf("Artist:%s", fileData.Meta.Artist))
	_ = WriteLine(osuFile, fmt.Sprintf("TitleUnicode:%s", fileData.Meta.Title))
	_ = WriteLine(osuFile, fmt.Sprintf("ArtistUnicode:%s", fileData.Meta.Artist))
	_ = WriteLine(osuFile, fmt.Sprintf("Creator:%s", AppendSubartistsToArtist(fileData.Meta.Artist, fileData.Meta.Subartists)))
	_ = WriteLine(osuFile, "Source:BMS")
	_ = WriteLine(osuFile, fmt.Sprintf("Tags:%s", fileData.Meta.Tags))
	_ = WriteLine(osuFile, fmt.Sprintf("Version:%s", GetDifficultyName(fileData.Meta.Difficulty, fileData.Meta.Subtitle)))
	_ = WriteLine(osuFile, "BeatmapID:0")
	_ = WriteLine(osuFile, "BeatmapSetID:0")

	_ = WriteLine(osuFile, "[Difficulty]")
	_ = WriteLine(osuFile, fmt.Sprintf("HPDrainRate:%.1f", conf.HPDrain))
	_ = WriteLine(osuFile, "CircleSize:8")
	_ = WriteLine(osuFile, fmt.Sprintf("OverallDifficulty:%.1f", conf.OverallDifficulty))
	_ = WriteLine(osuFile, "ApproachRate:0")
	_ = WriteLine(osuFile, "SliderMultiplier:1")
	_ = WriteLine(osuFile, "SliderTickRate:1")

	_ = WriteLine(osuFile, "[Events]")
	bg := fileData.Meta.StageFile
	if len(fileData.Meta.Banner) > 0 && len(fileData.Meta.StageFile) == 0 {
		bg = fileData.Meta.Banner
	}
	_ = WriteLine(osuFile, fmt.Sprintf("0,0,\"%s\",0,0", bg))

	if !conf.NoStoryboard {
		for i, bga := range fileData.BackgroundAnimation {
			endTime := 0.0
			if i+1 != len(fileData.BackgroundAnimation) {
				endTime = fileData.BackgroundAnimation[i+1].StartTime
			}
			vExt := path.Ext(bga.File)
			layer := "Background"
			if bga.Layer == Front {
				layer = "Foreground"
			}
			if !(vExt == ".wmv" || vExt == ".mpg" || vExt == ".avi" || vExt == ".mp4" || vExt == ".webm" || vExt == ".mkv") {
				_ = WriteLine(osuFile, fmt.Sprintf("Sprite,%s,%s,\"%s\",%d,%d", layer, "CentreRight", bga.File, 600, 240))
				// osu doesn't like decimals in starting/ending times
				_ = WriteLine(osuFile, fmt.Sprintf("_F,0,%d,%d,%d", int(bga.StartTime), int(endTime), 1))
			} else {
				_ = WriteLine(osuFile, fmt.Sprintf("Video,%d,\"%s\"", int(bga.StartTime), bga.File))
			}
		}
	}

	for _, sfx := range fileData.SoundEffects {
		_ = WriteLine(osuFile, fmt.Sprintf("Sample,%d,%d,\"%s\",%d", int(sfx.StartTime), 0, fileData.SoundStringArray[sfx.Sample-1], conf.Volume))
	}

	_ = WriteLine(osuFile, "[TimingPoints]")

	keys := make([]float64, len(fileData.TimingPoints))
	i := 0
	for k := range fileData.TimingPoints {
		keys[i] = k
		i++
	}
	sort.Float64s(keys)
	for j, k := range keys {
		if j == 0 {
			value := GetBeatDuration(fileData.TimingPoints[k])
			_ = WriteLine(osuFile, fmt.Sprintf("%f,%f,%d,%d,%d,%d,%d,%d", k, value, 4, 0, 0, conf.Volume, 1, 0))
		} else {
			beatDuration := GetBeatDuration(fileData.TimingPoints[k])
			if beatDuration == 0.0 {
				beatDuration = 999999999.0
			}
			// osu can't handle negative bpm(?)
			if beatDuration < 0.0 {
				beatDuration = math.Abs(beatDuration)
			}

			_ = WriteLine(osuFile, fmt.Sprintf("%f,%f,%d,%d,%d,%d,%d,%d", k, beatDuration, 4, 0, 0, conf.Volume, 1, 0))
			_ = WriteLine(osuFile, fmt.Sprintf("%f,%d,%d,%d,%d,%d,%d,%d", k, -100, 4, 0, 0, conf.Volume, 0, 0))
		}
		i++
	}

	_ = WriteLine(osuFile, "[HitObjects]")
	for lane, objects := range fileData.HitObjects {
		for _, obj := range objects {
			objType := 1 << 0
			if obj.IsLongNote {
				objType = 1 << 7
			}
			xPos := 64 * lane
			if lane == 8 {
				xPos = 32
			} else {
				xPos += 32
			}
			var hitSound string
			if obj.KeySounds != nil {
				hitSound = fileData.SoundStringArray[obj.KeySounds.Sample-1]
			}
			if objType == 1<<7 && int(obj.EndTime) > int(obj.StartTime) {
				_ = WriteLine(osuFile, fmt.Sprintf("%d,%d,%d,%d,%d,%d:0:0:0:0:%s",
					xPos,
					OsuYPos,
					int(obj.StartTime),
					objType,
					0,
					int(obj.EndTime),
					hitSound,
				))
			} else {
				_ = WriteLine(osuFile, fmt.Sprintf("%d,%d,%d,%d,%d,0:0:0:0:%s",
					xPos,
					OsuYPos,
					int(obj.StartTime),
					1<<0,
					0,
					hitSound,
				))
			}
		}
	}

	e = osuFile.Sync()
	if e != nil {
		return e
	}

	return nil
}
