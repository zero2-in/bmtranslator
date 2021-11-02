package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type JSONFileData struct {
	ProgramVersion string        `json:"program_version"`
	Version        string        `json:"version"`
	Metadata       BMSMetadata   `json:"metadata"`
	HitObjects     [][]HitObject `json:"hit_objects"`
	TimingPoints   []TimingPoint `json:"timing_points"`
	SampleIndex    []string      `json:"sample_index"`
	SoundEffects   []SoundEffect `json:"sound_effects"`
}

type TimingPoint struct {
	StartTime float64 `json:"start_time"`
	Bpm       float64 `json:"bpm"`
}

// ConvertBmsToJson outputs BMS information to a json file. Some information (like hex value mapping) is omitted.
func (conf *ProgramConfig) ConvertBmsToJson(fileData BMSFileData, outputPath string) error {
	dest, e := os.Create(outputPath)
	if e != nil {
		return e
	}

	defer dest.Close()

	d := &JSONFileData{
		Metadata:       fileData.Metadata,
		HitObjects:     make([][]HitObject, len(fileData.HitObjects)),
		TimingPoints:   make([]TimingPoint, 0),
		SampleIndex:    make([]string, 0),
		SoundEffects:   make([]SoundEffect, 0),
		Version:        JSONVersion,
		ProgramVersion: Version,
	}
	for i, h := range fileData.HitObjects {
		i--
		d.HitObjects[i] = make([]HitObject, len(fileData.HitObjects[i]))
		d.HitObjects[i] = h
	}
	for i, t := range fileData.TimingPoints {
		d.TimingPoints = append(d.TimingPoints, TimingPoint{
			StartTime: i,
			Bpm:       t,
		})
	}
	for _, s := range fileData.Audio.StringArray {
		d.SampleIndex = append(d.SampleIndex, s)
	}

	for _, s := range fileData.SoundEffects {
		d.SoundEffects = append(d.SoundEffects, s)
	}

	// Avoid null conflict
	//for i, h := range d.HitObjects {
	//	if len(h) == 0 {
	//		d.HitObjects[i] = make([]HitObject, 0)
	//	}
	//}

	j, e := json.Marshal(d)
	if e != nil {
		return e
	}

	e = ioutil.WriteFile(outputPath, j, 0644)
	if e != nil {
		return e
	}

	e = dest.Sync()
	if e != nil {
		return e
	}

	return nil
}
