package main

func CalculateTimingPoints(currentTime float64, startTrackWithBPM float64, data LocalTrackData) map[float64]float64 {
	// Create local points (k: time, v: bpm)
	points := map[float64]float64{}

	if len(data.BPMChanges) > 0 {
		timeElapsed := 0.0

		for i, tc := range data.BPMChanges {
			if i == 0 {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * tc.Position
			}
			stopTime := GetStopOffset(startTrackWithBPM, tc.Position, data)

			// Original BPM is already an absolute value. Here, it is safe to use
			// a negative for BPM because Quaver will automatically handle reverse scrolling, plus all
			// notes are already calculated based on an absolute value.
			b := tc.Bpm
			if tc.IsNegative {
				b = -tc.Bpm
			}
			points[currentTime+stopTime+timeElapsed] = b
			timeElapsed += GetBPMChangeOffset(i, data)
		}
	}

	if len(data.Stops) > 0 {
		// The time elapsed within the track, as we iterate through STOP commands.
		if len(data.BPMChanges) > 0 {
			timeElapsed := 0.0
			for i, tempoChange := range data.BPMChanges {
				if i == 0 {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * tempoChange.Position
				}
				// Iterate over all STOP commands to see if any lie within this tempo change.
				for _, stop := range data.Stops {
					if (i+1 < len(data.BPMChanges) && data.BPMChanges[i+1].Position > stop.Position && stop.Position >= tempoChange.Position) || i+1 == len(data.BPMChanges) {
						stopTime := GetStopOffset(startTrackWithBPM, stop.Position, data)
						// Adds the following: Time of beginning of track + time already passed by previous BPM changes
						// + time already passed by STOP commands + time passed based on location in range
						startAt := currentTime + timeElapsed + stopTime + (GetTrackDurationGivenBPM(tempoChange.Bpm, data.MeasureScale) * (stop.Position - tempoChange.Position))
						endAt := startAt + GetStopDuration(tempoChange.Bpm, stop.Duration)
						points[startAt] = 0.0
						points[endAt] = tempoChange.Bpm
					}
				}
				timeElapsed += GetBPMChangeOffset(i, data)
			}
		} else {
			// This block should run if there are no tempo changes in the entire track.
			timeElapsed := 0.0
			for stopIndex, stop := range data.Stops {
				if stopIndex == 0 {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * stop.Position
				}
				stopTime := GetStopOffset(startTrackWithBPM, stop.Position, data)
				points[currentTime+timeElapsed+stopTime] = 0.0
				points[currentTime+timeElapsed+stopTime+GetStopDuration(startTrackWithBPM, stop.Duration)] = startTrackWithBPM
				// Add additional time, so we know where we are in the next iteration.
				if stopIndex+1 < len(data.Stops) {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (data.Stops[stopIndex+1].Position - stop.Position)
				} else if stopIndex+1 == len(data.Stops) {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * ((1.0 - stop.Position) / 1.0)
				}
			}
		}
	}

	return points
}

func GetOffsetFromStartingTime(data *LocalTrackData, index int, message string, startTrackWithBPM float64) float64 {
	// Essentially beat snap
	measure := float64(len(message) / 2)
	notePos := float64(index) / measure

	// No change in tempo OR stop command. In this scenario, we can just ignore everything sent to us because it doesn't matter LOL
	if len(data.BPMChanges) == 0 && len(data.Stops) == 0 {
		return GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * notePos
	}

	// Hold the offset after the starting time.
	timeToAdd := 0.0
	// Try to find range of percentage the note belongs to.
	for i, t := range data.BPMChanges {
		if i == 0 {
			if notePos < t.Position {
				// Doesn't belong to any tempo change.
				// This is because it happens before the first tempo change.
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * notePos
				break
			} else {
				// Compensate for time before first tempo change
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * t.Position
			}
		}

		// If this is the last tempo change, OR current position is GREATER/EQUAL TO current tempo change
		if (i+1 == len(data.BPMChanges)) || (i+1 < len(data.BPMChanges) && data.BPMChanges[i+1].Position > notePos && notePos >= t.Position) {
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, data.MeasureScale) * (notePos - t.Position)
			break
		} else {
			// This tempo change happens too early before the note should be placed. Add its range.
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, data.MeasureScale) * (data.BPMChanges[i+1].Position - t.Position)
		}
	}
	if len(data.BPMChanges) == 0 {
		timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * notePos
	}
	if len(data.Stops) > 0 {
		timeToAdd += GetStopOffset(startTrackWithBPM, notePos, *data)
	}
	return timeToAdd
}
