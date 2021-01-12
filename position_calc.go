package main

func CalculateTimingPoints(currentTime float64, startTrackWithBPM float64, tempoChangeTimestampMap []LocalTempoChange, stopCommandMap []LocalStopCommand, scalar float64) map[float64]float64 {
	timingPoints := map[float64]float64{}

	if len(tempoChangeTimestampMap) > 0 {
		timeElapsed := 0.0

		for i, tc := range tempoChangeTimestampMap {
			if i == 0 {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * tc.Position
			}
			stopTime := GetStopOffset(startTrackWithBPM, tc.Position, tempoChangeTimestampMap, stopCommandMap, false)

			// Original BPM is already an absolute value. Here, it is safe to use
			// a negative for BPM because Quaver will automatically handle reverse scrolling, plus all
			// notes are already calculated based on an absolute value.
			b := tc.Bpm
			if tc.IsNegative {
				b = -tc.Bpm
			}
			timingPoints[currentTime+stopTime+timeElapsed] = b
			timeElapsed += GetTempoChangeOffset(i, scalar, tempoChangeTimestampMap)
		}
	}

	if len(stopCommandMap) > 0 {
		// The time elapsed within the track, as we iterate through STOP commands.
		if len(tempoChangeTimestampMap) > 0 {
			timeElapsed := 0.0
			for i, tempoChange := range tempoChangeTimestampMap {
				if i == 0 {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * tempoChange.Position
				}
				// Iterate over all STOP commands to see if any lie within this tempo change.
				for _, stop := range stopCommandMap {
					if (i+1 < len(tempoChangeTimestampMap) && tempoChangeTimestampMap[i+1].Position > stop.Position && stop.Position >= tempoChange.Position) || i+1 == len(tempoChangeTimestampMap) {
						stopTime := GetStopOffset(startTrackWithBPM, stop.Position, tempoChangeTimestampMap, stopCommandMap, false)
						// Adds the following: Time of beginning of track + time already passed by previous BPM changes
						// + time already passed by STOP commands + time passed based on location in range
						startAt := currentTime + timeElapsed + stopTime + (GetTrackDurationGivenBPM(tempoChange.Bpm, scalar) * (stop.Position - tempoChange.Position))
						endAt := startAt + GetStopDuration(tempoChange.Bpm, stop.Duration)
						timingPoints[startAt] = 0.0
						timingPoints[endAt] = tempoChange.Bpm
					}
				}
				timeElapsed += GetTempoChangeOffset(i, scalar, tempoChangeTimestampMap)
			}
		} else {
			// This block should run if there are no tempo changes in the entire track.
			timeElapsed := 0.0
			for stopIndex, stop := range stopCommandMap {
				if stopIndex == 0 {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * stop.Position
				}
				stopTime := GetStopOffset(startTrackWithBPM, stop.Position, tempoChangeTimestampMap, stopCommandMap, false)
				timingPoints[currentTime+timeElapsed+stopTime] = 0.0
				timingPoints[currentTime+timeElapsed+stopTime+GetStopDuration(startTrackWithBPM, stop.Duration)] = startTrackWithBPM

				// Add additional time, so we know where we are in the next iteration.
				if stopIndex+1 < len(stopCommandMap) {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * (stopCommandMap[stopIndex+1].Position - stop.Position)
				} else if stopIndex+1 == len(stopCommandMap) {
					timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * ((1.0 - stop.Position) / 1.0)
				}
			}
		}
	}

	return timingPoints
}

func GetOffsetFromStartingTime(startTrackWithBPM float64, tempoChangeTimestampMap []LocalTempoChange, stopCommandMap []LocalStopCommand, index int, message string, scalar float64) float64 {
	// Essentially beat snap
	measure := float64(len(message) / 2)
	notePos := float64(index) / measure

	// No change in tempo OR stop command. In this scenario, we can just ignore everything sent to us because it doesn't matter LOL
	if len(tempoChangeTimestampMap) == 0 && len(stopCommandMap) == 0 {
		return GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * notePos
	}

	// Hold the offset after the starting time.
	timeToAdd := 0.0
	// Try to find range of percentage the note belongs to.
	for i, t := range tempoChangeTimestampMap {
		if i == 0 {
			if notePos < t.Position {
				// Doesn't belong to any tempo change.
				// This is because it happens before the first tempo change.
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * notePos
				break
			} else {
				// Compensate for time before first tempo change
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * t.Position
			}
		}

		// If this is the last tempo change, OR current position is GREATER/EQUAL TO current tempo change
		if (i+1 == len(tempoChangeTimestampMap)) || (i+1 < len(tempoChangeTimestampMap) && tempoChangeTimestampMap[i+1].Position > notePos && notePos >= t.Position) {
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, scalar) * (notePos - t.Position)
			break
		} else {
			// This tempo change happens too early before the note should be placed. Add its range.
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, scalar) * (tempoChangeTimestampMap[i+1].Position - t.Position)
		}
	}
	if len(tempoChangeTimestampMap) == 0 {
		timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, scalar) * notePos
	}
	if len(stopCommandMap) > 0 {
		timeToAdd += GetStopOffset(startTrackWithBPM, notePos, tempoChangeTimestampMap, stopCommandMap, false)
	}
	return timeToAdd
}
