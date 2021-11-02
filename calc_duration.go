package main

// GetBeatDuration returns the duration of a single beat of
// a track, in 4/4 meter.
func GetBeatDuration(bpm float64) float64 {
	// Prevent division by 0
	if bpm == 0.0 {
		return 0.0
	}
	return (MinuteUnit / bpm) * Second
}

// getBaseTrackDuration returns the beat duration for the current BPM * 4.
// This is assuming tracks are in 4/4 meter. I am still not sure if BMS songs
// can be in a different meter.
func getBaseTrackDuration(currentBPM float64) float64 {
	return GetBeatDuration(currentBPM) * 4.0
}

// GetStopDuration gets the duration that the track should remain at 0 BPM for, based on the current BPM,
// and the duration. STOP commands are based on 1/192 of a whole note in 4/4.
func GetStopDuration(currentBPM float64, duration float64) float64 {
	return getBaseTrackDuration(currentBPM) * (duration / 192.0)
}

// GetTrackDurationGivenBPM gets the length of the track, in milliseconds, based on the BPM * the measure scale.
func GetTrackDurationGivenBPM(currentBPM float64, measureScale float64) float64 {
	return getBaseTrackDuration(currentBPM) * measureScale
}

// GetTotalTrackDuration gets the full length of the track. Different from GetTrackDurationGivenBPM,
// as this accounts for ALL BPM changes' offsets put together. It also considers the amount of time that
// is used for STOP directives.
func GetTotalTrackDuration(initialBPM float64, data LocalTrackData) float64 {
	var baseLength float64

	if len(data.BPMChanges) == 0 && len(data.Stops) == 0 {
		return GetTrackDurationGivenBPM(initialBPM, data.MeasureScale)
	}

	for i, change := range data.BPMChanges {
		if i == 0 {
			baseLength += GetTrackDurationGivenBPM(initialBPM, data.MeasureScale) * (change.Position / 100.0)
		}
		baseLength += GetBPMChangeOffset(i, data)
	}
	if len(data.BPMChanges) == 0 {
		baseLength += GetTrackDurationGivenBPM(initialBPM, data.MeasureScale)
	}
	stopTime := GetStopOffset(initialBPM, 100.0, data)
	return baseLength + stopTime
}
