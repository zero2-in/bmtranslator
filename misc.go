package main

import (
	"os"
	"regexp"
)

func GetDifficultyName(i string) string {
	return "Lv. " + i
}

func WriteLine(f *os.File, s string) error {
	_, e := f.WriteString(s + "\n")
	return e
}

func GetCorrespondingHitSound(hitSoundHexArray []string, target string, volume int) *HitObjectKeySound {
	for ind, v := range hitSoundHexArray {
		if v == target {
			return &HitObjectKeySound{
				Volume: volume,
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
