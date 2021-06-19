package main

// GetStopOffset gets the total amount of time that all STOP before the current time would cause.
func GetStopOffset(initialBPM float64, pos float64, data LocalTrackData) float64 {
	totalOffset := 0.0

	// Return 0 if there are no stop commands
	if len(data.Stops) == 0 {
		return totalOffset
	}

	for _, stop := range data.Stops {
		if pos > stop.Position {
			bpmToUse := initialBPM
			for i, change := range data.BPMChanges {
				if (i+1 < len(data.BPMChanges) && data.BPMChanges[i+1].Position > stop.Position && stop.Position >= change.Position) || (i+1 == len(data.BPMChanges) && stop.Position >= change.Position) {
					bpmToUse = change.Bpm
					break
				}
			}

			totalOffset += GetStopDuration(bpmToUse, stop.Duration)
		}
	}
	return totalOffset
}

// Depending on where the index lies, calculate how much time to add
// based on the BPM changes given for the track.
func GetBPMChangeOffset(currentIndex int, data LocalTrackData) float64 {
	if len(data.BPMChanges) == 0 {
		return 0.0
	}
	if currentIndex+1 < len(data.BPMChanges) {
		return GetTrackDurationGivenBPM(data.BPMChanges[currentIndex].Bpm, data.MeasureScale) * ((data.BPMChanges[currentIndex+1].Position - data.BPMChanges[currentIndex].Position) / 100.0)
	} else if currentIndex+1 == len(data.BPMChanges) {
		return GetTrackDurationGivenBPM(data.BPMChanges[currentIndex].Bpm, data.MeasureScale) * ((100.0 - data.BPMChanges[currentIndex].Position) / 100.0)
	}
	return 0.0
}
