package main

import (
	"github.com/fatih/color"
	"regexp"
	"sort"
	"strings"
)

var (
	//ignore mines....for now.
	// Quaver might implement this in the future
	//mineRegex        = regexp.MustCompile("[d][1-9]")
	noteRegex        = regexp.MustCompile("[1][1-9]")
	player2NoteRegex = regexp.MustCompile("[2][1-9]")
	lnRegex          = regexp.MustCompile("[5][1-z]")
	player2LnRegex   = regexp.MustCompile("[6][1-9]")
)

// ReadFileData converts from BMS to a ConvertedFile. Returns a ConvertedFile, whether file was skipped or not, and an error if it errored.
func (conf *ProgramConfig) ReadFileData(inputPath string, bmsFileName string) (*FileData, error) {

	// What time (ms) the current track will start at.
	startTrackAt := 0.0

	// What BPM the current track will start at.
	var startTrackWithBPM float64

	// Tracks long notes in channels 51-59 when they start.
	// Normal LNs (that use #LNOBJ) will not need this.
	longNoteTracker := map[int]float64{}
	longNoteSoundEffectTracker := map[int]*KeySound{}

	fileData, e := conf.CompileBMSToStruct(inputPath, bmsFileName)
	if e != nil {
		return nil, e
	}
	if fileData == nil {
		return nil, nil
	}
	startTrackWithBPM = fileData.StartingBPM

	fileData.TimingPoints[0.0] = fileData.StartingBPM

	// Sort all tracks in ascending order
	keys := make([]int, 0)
	for k := range fileData.TrackLines {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, trackInt := range keys {
		localTrackData, e := conf.ReadTrackData(trackInt, fileData.TrackLines[trackInt], fileData.BPMChangeIndex, fileData.StopIndex)
		if e != nil {
			return nil, e
		}
		if localTrackData == nil {
			return nil, nil
		}

		for _, line := range fileData.TrackLines[trackInt] {
			if len(line.Message)%2 != 0 {
				continue
			}

			// Cancel parsing if notes are found in P2 side.
			if player2NoteRegex.MatchString(line.Channel) || player2LnRegex.MatchString(line.Channel) {
				if conf.Verbose {
					color.HiYellow("* This map has notes in player 2's side, which would overlap player 1. Not going to process this map.")
					return nil, nil
				}
			}
			if !(noteRegex.MatchString(line.Channel) || lnRegex.MatchString(line.Channel) || line.Channel == "01" || line.Channel == "04" || line.Channel == "07") {
				continue
			}
			for i := 0; i < len(line.Message)/2; i++ {
				if (i*2)+2 > len(line.Message) {
					break
				}
				target := getHexadecimalPair(i, line.Message)
				if target == "00" {
					continue
				}
				localOffset := GetOffsetFromStartingTime(localTrackData, i, line.Message, startTrackWithBPM)
				sfx := conf.GetCorrespondingHitSound(fileData.SoundHexArray, target)
				laneInt := strings.Index(Base36Range, line.Channel[1:])
				// maybe you should get among some bitches
				if noteRegex.MatchString(line.Channel) || lnRegex.MatchString(line.Channel) {
					if (laneInt == 6 && !conf.NoScratchLane) || laneInt != 6 {
						if laneInt == 6 {

							// Uses the channel for the scratch lane, manually adjust to lane 8
							laneInt = 8
						} else if laneInt >= 8 {
							// Compensate for notes past 8th key (6th and 7th lane)
							laneInt -= 2
						}
						if laneInt > 8 {
							color.HiRed("* File wants more than 8 keys, skipping")
							return nil, nil
						}
						hitObject := HitObject{
							StartTime: startTrackAt + localOffset,
						}

						// Closes the long note
						if target == fileData.LnObject {
							if len(fileData.HitObjects[laneInt]) == 0 {
								// Why is there an LN tail as the first object??
								continue
							}
							back := len(fileData.HitObjects[laneInt]) - 1
							//if fileData.HitObjects[back].KeySounds == nil {
							//	// Previous hit object didn't have key sounds.
							//	// That means the previous value is a LN object, so we don't add a new one.
							//	continue
							//}
							// If the LN is too short don't actually use it.
							if hitObject.StartTime-fileData.HitObjects[laneInt][back].StartTime < 2.0 {
								continue
							}
							fileData.HitObjects[laneInt][back].IsLongNote = true
							fileData.HitObjects[laneInt][back].EndTime = hitObject.StartTime
							continue
						}
						if sfx != nil {
							hitObject.KeySounds = sfx
						}

						// This is a long note existing in channels 51-59. We save it to a map storing these values.
						if lnRegex.MatchString(line.Channel) {
							// This is the end of a long note. Now, we can place the note.
							if longNoteTracker[laneInt] != 0.0 {
								// haha funny end time joke
								hitObject.EndTime = hitObject.StartTime
								hitObject.StartTime = longNoteTracker[laneInt]
								hitObject.IsLongNote = true
								if longNoteSoundEffectTracker[laneInt] != nil {
									hitObject.KeySounds = &KeySound{
										Sample: longNoteSoundEffectTracker[laneInt].Sample,
										Volume: longNoteSoundEffectTracker[laneInt].Volume,
									}
								}
								// Reset values
								longNoteTracker[laneInt] = 0.0
								longNoteSoundEffectTracker[laneInt] = nil
								// Invalid long note because it ends either before or exactly at the position it ends.
								// In other words, do not process it.
								if hitObject.EndTime <= hitObject.StartTime {
									continue
								}
							} else {
								// This is the head of a long note, so we store its start time and key sounds for later.
								longNoteTracker[laneInt] = hitObject.StartTime
								longNoteSoundEffectTracker[laneInt] = hitObject.KeySounds
								continue
							}
						}
						fileData.HitObjects[laneInt] = append(fileData.HitObjects[laneInt], hitObject)
						continue
					}
				}
				if line.Channel == "01" || laneInt == 6 && conf.NoScratchLane {
					// Sound effect (channel 01)
					soundEffect := SoundEffect{
						StartTime: startTrackAt + localOffset,
					}
					if sfx != nil {
						soundEffect.Sample = sfx.Sample
						soundEffect.Volume = sfx.Volume
					} else {
						// No sound effect assigned to this address?
						continue
					}
					fileData.SoundEffects = append(fileData.SoundEffects, soundEffect)
				}
				if (line.Channel == "04" || line.Channel == "07") && conf.FileType == Osu && !conf.NoStoryboard {
					t := fileData.BGAIndex[target]
					l := Back
					if line.Channel == "07" {
						l = Front
					}
					if len(t) > 0 {
						fileData.BackgroundAnimation = append(fileData.BackgroundAnimation, BackgroundAnimation{
							StartTime: startTrackAt + localOffset,
							File:      t,
							Layer:     l,
						})
					}
				}
			}
		}

		// Get the full length of the track.
		fullLengthOfTrack := GetTotalTrackDuration(startTrackWithBPM, *localTrackData)
		// Calculate all timing points.
		if !conf.NoTimingPoints {
			timingPoints := CalculateTimingPoints(startTrackAt, startTrackWithBPM, *localTrackData)
			for k, v := range timingPoints {
				fileData.TimingPoints[k] = v
			}
		}
		if len(localTrackData.BPMChanges) > 0 {
			startTrackWithBPM = localTrackData.BPMChanges[len(localTrackData.BPMChanges)-1].Bpm
		}

		// Add the length of the track onto the current time.
		startTrackAt += fullLengthOfTrack

		if !conf.NoMeasureLines && !conf.NoTimingPoints {
			fileData.TimingPoints[startTrackAt] = startTrackWithBPM
		}
	}

	sort.Slice(fileData.BackgroundAnimation, func(i, j int) bool {
		return fileData.BackgroundAnimation[i].StartTime < fileData.BackgroundAnimation[j].StartTime
	})

	return fileData, nil
}
