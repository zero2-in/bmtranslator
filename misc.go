package main

import (
	"os"
	"regexp"
	"strings"
)

func GetDifficultyName(i string, sub string) string {
	b := "Lv. " + i
	if len(sub) == 0 {
		return b
	}
	return sub + " " + b
}

func AppendSubartistsToArtist(a string, subartists []string) string {
	if len(subartists) == 0 {
		return a
	}
	return a + " <" + strings.Join(subartists, " | ") + ">"
}

func WriteLine(f *os.File, s string) error {
	_, e := f.WriteString(s + "\n")
	return e
}

func (conf *ProgramConfig) GetCorrespondingHitSound(hitSoundHexArray []string, target string) *KeySound {
	for ind, v := range hitSoundHexArray {
		if v == target {
			return &KeySound{
				Volume: conf.Volume,
				Sample: ind + 1,
			}
		}
	}
	return nil
}

func RegSplit(text string, delimeter string) []string {
	reg := regexp.MustCompile(delimeter)
	indexes := reg.FindAllStringIndex(text, -1)
	laststart := 0
	result := make([]string, len(indexes)+1)
	for i, element := range indexes {
		result[i] = text[laststart:element[0]]
		laststart = element[1]
	}
	result[len(indexes)] = text[laststart:]
	return result
}

func ClampInt(i int, max int, min int) int {
	if i > max {
		return max
	}
	if i < min {
		return min
	}
	return i
}

func ClampFloat(i float64, max float64, min float64) float64 {
	if i > max {
		return max
	}
	if i < min {
		return min
	}
	return i
}
