package main

import (
	"fmt"
	"math"
	"os"
	"path"
	"sort"
	"strconv"
)

const (
	OsuYPos = 192
)

// Convert a BMS file to .osu (for the game osu!).
func ConvertBmsToOsu(convertedFile ConvertedFile, outputPath string, hpDrain float64, overallDifficulty float64, volume int) error {
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
	_ = WriteLine(osuFile, "TitleUnicode:"+convertedFile.Metadata.Title)
	_ = WriteLine(osuFile, "ArtistUnicode:"+convertedFile.Metadata.Artist)
	_ = WriteLine(osuFile, "Creator:"+convertedFile.Metadata.Creator)
	_ = WriteLine(osuFile, "Source:BMS")
	_ = WriteLine(osuFile, "Tags:"+convertedFile.Metadata.Tags)
	_ = WriteLine(osuFile, "Version:"+GetDifficultyName(convertedFile.Metadata.Difficulty))
	_ = WriteLine(osuFile, "BeatmapID:0")
	_ = WriteLine(osuFile, "BeatmapSetID:0")

	_ = WriteLine(osuFile, "[Difficulty]")
	_ = WriteLine(osuFile, fmt.Sprintf("HPDrainRate:%.1f", hpDrain))
	_ = WriteLine(osuFile, "CircleSize:8")
	_ = WriteLine(osuFile, fmt.Sprintf("OverallDifficulty:%.1f", overallDifficulty))
	_ = WriteLine(osuFile, "ApproachRate:0")
	_ = WriteLine(osuFile, "SliderMultiplier:1")
	_ = WriteLine(osuFile, "SliderTickRate:1")

	_ = WriteLine(osuFile, "[Events]")
	_ = WriteLine(osuFile, fmt.Sprintf("0,0,\"%s\",0,0", convertedFile.Metadata.StageFile))

	for i, bga := range convertedFile.BackgroundAnimation {
		endTime := ""
		if i+1 != len(convertedFile.BackgroundAnimation) {
			endTime = strconv.Itoa(int(convertedFile.BackgroundAnimation[i+1].StartTime))
		}
		vExt := path.Ext(bga.File)
		layer := "Background"
		if bga.Layer == Front {
			layer = "Foreground"
		}
		if !(vExt == ".wmv" || vExt == ".mpg" || vExt == ".avi" || vExt == ".mp4" || vExt == ".webm" || vExt == ".mkv") {
			_ = WriteLine(osuFile, fmt.Sprintf("Sprite,%s,%s,\"%s\",%d,%d", layer, "CentreRight", bga.File, 600, 240))
			// osu doesn't like decimals in starting/ending times
			_ = WriteLine(osuFile, fmt.Sprintf("_F,0,%d,%s,%d", int(bga.StartTime), endTime, 1))
		} else {
			_ = WriteLine(osuFile, fmt.Sprintf("Video,%d,\"%s\"", int(bga.StartTime), bga.File))
		}
	}
	for _, sfx := range convertedFile.SoundEffects {
		_ = WriteLine(osuFile, fmt.Sprintf("Sample,%d,%d,\"%s\",%d", int(sfx.StartTime), 0, convertedFile.KeySoundStringArray[sfx.Sample-1], volume))
	}

	_ = WriteLine(osuFile, "[TimingPoints]")

	keys := make([]float64, len(convertedFile.TimingPoints))
	i := 0
	for k := range convertedFile.TimingPoints {
		keys[i] = k
		i++
	}
	sort.Float64s(keys)
	for j, k := range keys {
		if j == 0 {
			value := GetBeatDuration(convertedFile.TimingPoints[k])
			_ = WriteLine(osuFile, fmt.Sprintf("%f,%f,%d,%d,%d,%d,%d,%d", k, value, 4, 0, 0, volume, 1, 0))
		} else {
			beatDuration := GetBeatDuration(convertedFile.TimingPoints[k])
			if beatDuration == 0.0 {
				beatDuration = 999999999.0
			}
			// osu can't handle negative bpm(?)
			if beatDuration < 0.0 {
				beatDuration = math.Abs(beatDuration)
			}

			_ = WriteLine(osuFile, fmt.Sprintf("%f,%f,%d,%d,%d,%d,%d,%d", k, beatDuration, 4, 0, 0, volume, 1, 0))
			_ = WriteLine(osuFile, fmt.Sprintf("%f,%d,%d,%d,%d,%d,%d,%d", k, -100, 4, 0, 0, volume, 0, 0))
		}
		i++
	}

	_ = WriteLine(osuFile, "[HitObjects]")
	for _, obj := range convertedFile.HitObjects {
		objType := 1 << 0
		if obj.IsLongNote {
			objType = 1 << 7
		}
		xPos := 64 * obj.Lane
		if obj.Lane == 8 {
			xPos = 32
		} else {
			xPos += 32
		}
		var hitSound string
		if obj.KeySounds != nil {
			hitSound = convertedFile.KeySoundStringArray[obj.KeySounds.Sample-1]
		}
		if objType == 1<<7 {
			// x,y (0),time,type,hitSound,endTime
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
				objType,
				0,
				hitSound,
			))
		}
	}

	e = osuFile.Sync()
	if e != nil {
		return e
	}

	return nil
}
