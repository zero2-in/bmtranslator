package main

import (
	"fmt"
	"os"
	"strings"
)

const (
	DumpVersion = "0.0.1"
)

// Dump file data to a txt file.
func (conf *ProgramConfig) WriteDump(fileData FileData, outputPath string, forBmsFile string) error {
	dest, e := os.Create(outputPath)
	if e != nil {
		return e
	}
	defer dest.Close()

	// flush contents to osu
	_ = WriteLine(dest, "##### BMTRANSLATOR FILEDATA DUMP #####")
	_ = WriteLine(dest, fmt.Sprintf("for %s", forBmsFile))
	_ = WriteLine(dest, fmt.Sprintf("version %s", DumpVersion))

	_ = WriteLine(dest, fmt.Sprintf("title %s", fileData.Meta.Title))
	_ = WriteLine(dest, fmt.Sprintf("artist %s", fileData.Meta.Artist))
	_ = WriteLine(dest, fmt.Sprintf("tags %s", fileData.Meta.Tags))
	_ = WriteLine(dest, fmt.Sprintf("difficulty %s", fileData.Meta.Difficulty))
	_ = WriteLine(dest, fmt.Sprintf("stagefile %s", fileData.Meta.StageFile))
	_ = WriteLine(dest, fmt.Sprintf("subtitle %s", fileData.Meta.Subtitle))
	_ = WriteLine(dest, fmt.Sprintf("subartists %s", strings.Join(fileData.Meta.Subartists, ",")))

	_ = WriteLine(dest, fmt.Sprintf("bpm %f", fileData.StartingBPM))
	_ = WriteLine(dest, fmt.Sprintf("lnobject %s", fileData.LnObject))
	_ = WriteLine(dest, "sound index")
	for i, v := range fileData.SoundHexArray {
		_ = WriteLine(dest, fmt.Sprintf("%s %s", v, fileData.SoundStringArray[i]))
	}
	_ = WriteLine(dest, "bpm change index")
	for k, v := range fileData.BPMChangeIndex {
		_ = WriteLine(dest, fmt.Sprintf("%s %f", k, v))
	}
	_ = WriteLine(dest, "stop index")
	for k, v := range fileData.StopIndex {
		_ = WriteLine(dest, fmt.Sprintf("%s %f", k, v))
	}
	_ = WriteLine(dest, "bga index")
	for k, v := range fileData.BGAIndex {
		_ = WriteLine(dest, fmt.Sprintf("%s %s", k, v))
	}
	_ = WriteLine(dest, "sound effects")
	for _, v := range fileData.SoundEffects {
		_ = WriteLine(dest, fmt.Sprintf("%f %s", v.StartTime, fileData.SoundHexArray[v.Sample-1]))
	}
	_ = WriteLine(dest, "hit objects")
	for _, v := range fileData.HitObjects {
		isLn := 0
		if v.IsLongNote {
			isLn = 1
		}
		var keysound string
		if v.KeySounds == nil {
			keysound = "null"
		} else {
			keysound = fileData.SoundHexArray[v.KeySounds.Sample-1]
		}
		_ = WriteLine(dest, fmt.Sprintf("%f %d %d %f %s", v.StartTime, v.Lane, isLn, v.EndTime, keysound))
	}
	_ = WriteLine(dest, "timing points")
	for k, v := range fileData.TimingPoints {
		_ = WriteLine(dest, fmt.Sprintf("%f %f", k, v))
	}
	_ = WriteLine(dest, "background animation")
	for _, v := range fileData.BackgroundAnimation {
		l := 0
		if v.Layer == Front {
			l = 1
		}
		_ = WriteLine(dest, fmt.Sprintf("%f %d %s", v.StartTime, l, v.File))
	}

	e = dest.Sync()
	if e != nil {
		return e
	}

	return nil
}
