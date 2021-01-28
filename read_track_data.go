package main

import (
	"github.com/fatih/color"
	"sort"
	"strconv"
)

func (conf *ProgramConfig) ReadTrackData(trackNumber int, lines []Line, bpmChangeIndex map[string]float64, stopIndex map[string]float64) (*LocalTrackData, error) {
	localTrackData := &LocalTrackData{
		MeasureScale: 1.0,
		Stops:        make([]LocalStop, 0),
		BPMChanges:   make([]LocalBPMChange, 0),
	}

	for _, line := range lines {
		switch line.Channel {
		case "02":
			i, e := strconv.ParseFloat(line.Message, 64)
			if e != nil {
				if conf.Verbose {
					color.HiRed("* Measure scale is invalid. cannot continue parsing (Track: %d)", trackNumber)
				}
				return nil, nil
			}
			if i <= 0.0 {
				if conf.Verbose {
					color.HiRed("* Measure scale is negative or 0. cannot continue parsing (Track: %d)", trackNumber)
				}
				return nil, nil
			}
			localTrackData.MeasureScale = i
			continue
		case "08", "03":
			// If either 08 or 03 has no message,
			// most implementations choose to read this as 0 bpm.
			if len(line.Message) == 0 {
				continue
			}
			for i := 0; i < len(line.Message)/2; i++ {
				if getHexadecimalPair(i, line.Message) == "00" {
					continue
				}
				var bpm float64
				if line.Channel == "03" {
					// 03 is used for hexadecimal BPM changes from 1-255.
					parsedBpm, e := strconv.ParseInt(getHexadecimalPair(i, line.Message), 16, 64)
					if e != nil {
						// On error, consider the BPM to be 0.
						bpm = 0.0
					} else {
						bpm = float64(parsedBpm)
					}
				} else {
					// will automatically be 0 (f64) if not found because of the way go works
					bpm = bpmChangeIndex[getHexadecimalPair(i, line.Message)]
				}

				// associate % in track with bpm change
				localTrackData.BPMChanges = append(localTrackData.BPMChanges, LocalBPMChange{
					Position:   getFraction(i, line.Message),
					Bpm:        bpm,
					IsNegative: bpm < 0.0,
				})
			}
			continue
		case "09":
			if len(line.Message) == 0 {
				continue
			}
			for i := 0; i < len(line.Message)/2; i++ {
				if getHexadecimalPair(i, line.Message) == "00" {
					continue
				}
				// True if a STOP command for said message part appears in the known STOP command lengths.
				// if not, it was probably invalid.
				if val, ok := stopIndex[getHexadecimalPair(i, line.Message)]; ok {
					localTrackData.Stops = append(localTrackData.Stops, LocalStop{
						Position: getFraction(i, line.Message),
						Duration: val,
					})
				}
			}
			continue
		}
	}

	sort.Slice(localTrackData.BPMChanges, func(i, j int) bool {
		return localTrackData.BPMChanges[i].Position < localTrackData.BPMChanges[j].Position
	})
	sort.Slice(localTrackData.Stops, func(i, j int) bool {
		return localTrackData.Stops[i].Position < localTrackData.Stops[j].Position
	})

	return localTrackData, nil
}
