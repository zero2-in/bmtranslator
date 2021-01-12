package main

func GetBeatDuration(bpm float64) float64 {
	// Prevent division by 0
	if bpm == 0.0 {
		return 0.0
	}
	return (MinuteUnit / bpm) * Millisecond
}

func getBaseTrackDuration(currentBpm float64) float64 {
	return GetBeatDuration(currentBpm) * 4.0
}

// floating point hell
func GetStopDuration(currentBpm float64, duration float64) float64 {
	return getBaseTrackDuration(currentBpm) * (duration / 192.0)
}

func GetTrackDurationGivenBPM(currentBpm float64, scalar float64) float64 {
	return getBaseTrackDuration(currentBpm) * scalar
}

// Gets the full length of the track. Different from GetTrackDurationGivenBPM, as this accounts for ALL tempo changes' offsets put together.
func GetTotalTrackDuration(initialBPM float64, tempoChanges []LocalTempoChange, stopCommands []LocalStopCommand, scalar float64) float64 {
	baseLength := 0.0

	if len(tempoChanges) == 0 && len(stopCommands) == 0 {
		return GetTrackDurationGivenBPM(initialBPM, scalar)
	}

	for i, tc := range tempoChanges {
		if i == 0 {
			baseLength += GetTrackDurationGivenBPM(initialBPM, scalar) * tc.Position
		}
		baseLength += GetTempoChangeOffset(i, scalar, tempoChanges)
	}
	if len(tempoChanges) == 0 {
		baseLength += GetTrackDurationGivenBPM(initialBPM, scalar)
	}
	// No stop commands? Return the base length since the track length has already been found.
	stopTime := GetStopOffset(initialBPM, 1.0, tempoChanges, stopCommands, false)
	return baseLength + stopTime
}
