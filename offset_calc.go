package main

// Gets the total amount of time that all STOP offsets before the current position would cause.
// explicitEqual is deprecated and also cuz im too lazy to remove it lol
func GetStopOffset(initialBPM float64, pos float64, tempoChangeTimestampMap []LocalTempoChange, stopCommandMap []LocalStopCommand, explicitEqual bool) float64 {
	totalOffset := 0.0

	// Return 0 if there are no stop commands
	if len(stopCommandMap) == 0 {
		return totalOffset
	}

	for _, stop := range stopCommandMap {
		if pos > stop.Position {
			bpmToUse := initialBPM
			for i, tempoChange := range tempoChangeTimestampMap {
				if (i+1 < len(tempoChangeTimestampMap) && tempoChangeTimestampMap[i+1].Position > stop.Position && stop.Position >= tempoChange.Position) || i+1 == len(tempoChangeTimestampMap) {
					bpmToUse = tempoChange.Bpm
					break
				}
			}
			totalOffset += GetStopDuration(bpmToUse, stop.Duration)
		}
	}
	return totalOffset
}

// Depending on where the index lies, calculate how much time to add
// based on the tempo changes given for the track.
func GetTempoChangeOffset(currentIndex int, scalar float64, tempoChanges []LocalTempoChange) float64 {
	if len(tempoChanges) == 0 {
		return 0.0
	}
	if currentIndex+1 < len(tempoChanges) {
		return GetTrackDurationGivenBPM(tempoChanges[currentIndex].Bpm, scalar) * (tempoChanges[currentIndex+1].Position - tempoChanges[currentIndex].Position)
	} else if currentIndex+1 == len(tempoChanges) {
		return GetTrackDurationGivenBPM(tempoChanges[currentIndex].Bpm, scalar) * ((1.0 - tempoChanges[currentIndex].Position) / 1.0)
	}
	return 0.0
}
