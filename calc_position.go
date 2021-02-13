package main

func CalculateTimingPoints(currentTime float64, startTrackWithBPM float64, data LocalTrackData) map[float64]float64 {
	// Create local points (k: time, v: bpm)
	points := map[float64]float64{}

	if len(data.BPMChanges) > 0 {
		timeElapsed := 0.0

		for i, tc := range data.BPMChanges {
			if i == 0 {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (tc.Position / 100.0)
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

	// scuffed
	if len(data.Stops) > 0 {
		timeElapsed := 0.0
		for stopIndex, stop := range data.Stops {
			if len(data.BPMChanges) > 0 {
				localTimeElapsed := 0.0
				stopTime := GetStopOffset(startTrackWithBPM, stop.Position, data)
				// Search BPM changes and find out if a STOP is located within the range of any.
				for i, bpmChange := range data.BPMChanges {
					if i == 0 {
						localTimeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (bpmChange.Position / 100.0)
					}
					if (i+1 < len(data.BPMChanges) && data.BPMChanges[i+1].Position > stop.Position && stop.Position >= bpmChange.Position) || (i+1 == len(data.BPMChanges) && stop.Position >= bpmChange.Position) {
						// Adds the following: Time of beginning of track + time already passed by previous BPM changes
						// + time already passed by STOP commands + time passed based on location in range
						startAt := currentTime + localTimeElapsed + stopTime + (GetTrackDurationGivenBPM(bpmChange.Bpm, data.MeasureScale) * ((stop.Position - bpmChange.Position) / 100.0))
						endAt := startAt + GetStopDuration(bpmChange.Bpm, stop.Duration)

						points[startAt] = 0.0
						points[endAt] = bpmChange.Bpm
						break
					} else if i+1 == len(data.BPMChanges) && stop.Position < data.BPMChanges[0].Position {
						startAt := currentTime + stopTime + (GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (stop.Position / 100.0))
						endAt := startAt + GetStopDuration(startTrackWithBPM, stop.Duration)
						points[startAt] = 0.0
						points[endAt] = startTrackWithBPM
						break
					}
					localTimeElapsed += GetBPMChangeOffset(i, data)
				}
				continue
			}

			if stopIndex == 0 {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (stop.Position / 100.0)
			}

			stopTime := GetStopOffset(startTrackWithBPM, stop.Position, data)
			points[currentTime+timeElapsed+stopTime] = 0.0
			points[currentTime+timeElapsed+stopTime+GetStopDuration(startTrackWithBPM, stop.Duration)] = startTrackWithBPM
			if stopIndex+1 < len(data.Stops) {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * ((data.Stops[stopIndex+1].Position - stop.Position) / 100.0)
			} else if stopIndex+1 == len(data.Stops) {
				timeElapsed += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * ((100.0 - stop.Position) / 100.0)
			}
		}
	}

	return points
}

func GetOffsetFromStartingTime(data *LocalTrackData, index int, message string, startTrackWithBPM float64) float64 {
	// Essentially beat snap
	measure := float64(len(message) / 2)
	notePos := (float64(index) / measure) * 100.0

	// No change in BPM OR stop command. In this scenario, we can just ignore everything sent to us because it doesn't matter LOL
	if len(data.BPMChanges) == 0 && len(data.Stops) == 0 {
		return GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (notePos / 100.0)
	}

	timeToAdd := 0.0
	for i, t := range data.BPMChanges {
		if i == 0 {
			if notePos < t.Position {
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (notePos / 100.0)
				break
			} else {
				timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (t.Position / 100.0)
			}
		}

		// search if this note belongs to a certain range.
		// explanation of this specific part:
		// if there's one bpm change/it's the last bpm change, and the note position is greater than or equal to it OR:
		// there is more than one bpm change but not the last one, and the position of the next bpm change is greater than the note position and the not position is greater than or equal to the bpm change position.
		if ((i+1 == len(data.BPMChanges)) && notePos >= t.Position) || (i+1 < len(data.BPMChanges) && data.BPMChanges[i+1].Position > notePos && notePos >= t.Position) {
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, data.MeasureScale) * ((notePos - t.Position) / 100.0)
			break
		} else if i+1 < len(data.BPMChanges) {
			// if the note didn't match any of the above criteria,
			// and there's still more bpm changes left, add on more time.
			timeToAdd += GetTrackDurationGivenBPM(t.Bpm, data.MeasureScale) * ((data.BPMChanges[i+1].Position - t.Position) / 100.0)
		}
	}
	if len(data.BPMChanges) == 0 {
		timeToAdd += GetTrackDurationGivenBPM(startTrackWithBPM, data.MeasureScale) * (notePos / 100.0)
	}
	if len(data.Stops) > 0 {
		timeToAdd += GetStopOffset(startTrackWithBPM, notePos, *data)
	}
	return timeToAdd
}
