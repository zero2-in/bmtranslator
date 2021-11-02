package main

import (
	"os"
	"strings"
)

func getHexadecimalPair(i int, str string) string {
	return str[i*2 : (i*2)+2]
}

func getFraction(i int, str string) float64 {
	return 100.0 * (float64(i) / (float64(len(str)) / 2.0))
}

func GetDifficultyName(i string, sub string, autoScratch bool) string {
	if i == "0" {
		i = "Special"
	}
	b := "Lv. " + i
	if autoScratch {
		b = "[Auto Scratch] Lv. " + i
	}
	if len(sub) == 0 {
		return b
	}
	return sub + " " + b
}

func AppendSubArtistsToArtist(a string, subartists []string) string {
	if len(subartists) == 0 {
		return a
	}
	return a + " <" + strings.Join(subartists, " | ") + ">"
}

func WriteLine(f *os.File, s string) error {
	_, e := f.WriteString(s + "\n")
	return e
}

// GetCorrespondingHitSound gets a hexadecimal value's hit sound as a sample index for future reference.
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
