package main

import (
	"bufio"
	"github.com/fatih/color"
	"math"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	mineRegex        = regexp.MustCompile("[d-e][1-9]")
	noteRegex        = regexp.MustCompile("[1][1-9]")
	player2NoteRegex = regexp.MustCompile("[2][1-9]")
	lnRegex          = regexp.MustCompile("[5-6][1-z]")
	bracketRegex     = regexp.MustCompile("\\[([^]]+)]")
)

// Converts from BMS to a ConvertedFile. Returns a ConvertedFile, whether file was skipped or not, and an error if it errored.
func GetConvertedFile(inputPath string, bmsFileName string, verbose bool, volume int, noBracketsInTitle bool) (*ConvertedFile, error) {
	file, err := os.Open(path.Join(inputPath, bmsFileName))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	convertedFile := ConvertedFile{
		Metadata: BMSMetadata{
			Artist:     "Unknown artist",
			Difficulty: "--",
		},
		TimingPoints: map[float64]float64{},
	}

	// Store data about ALL track data within a 2D slice.
	TrackData := map[int][]LocalTrackData{}

	// The BPM to start with. It can be changed.
	// This is the value that a new track will start with based on the last track's final BPM command, if it had one.
	// BM98 default is 130.
	startTrackWithBPM := 130.0

	// What time the current track will start at. very important in keeping track of where to place notes.
	currentTime := 0.0

	// Tracks long notes when they start.
	longNoteTracker := map[int]float64{}
	longNoteSoundEffectTracker := map[int]*HitObjectKeySound{}

	var hitSoundHexArray []string
	bpmChanges := map[string]float64{}
	stopCommands := map[string]float64{}
	bga := map[string]string{}

	lnObject := ""

	// Aggregate data into TrackData map
	lineIndex := 1
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "#") {
			continue
		}
		lineLower := strings.ToLower(line)

		if len(line) < 7 || line[6] != ':' {
			if strings.HasPrefix(lineLower, "#player") {
				if len(line) < 9 {
					if verbose {
						color.HiRed("* Player type cannot be determined. (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				switch line[8] {
				case '1':
					break
				case '2':
					if verbose {
						color.HiRed("* This is for player 2; this is not supported")
					}
					return nil, nil
				case '3':
					if verbose {
						color.HiRed("* Double play mode; this is not supported")
					}
					return nil, nil
				default:
					if verbose {
						color.HiRed("* Even though player header was defined, there was no valid input (Line: %d)", lineIndex)
					}
					return nil, nil
				}
			} else if strings.HasPrefix(lineLower, "#genre") {
				if len(line) < 8 {
					if verbose {
						color.HiYellow("* #genre is invalid, ignoring (Line: %d)", lineIndex)
					}
					convertedFile.Metadata.Tags = "BMT, BMS"
					continue
				}
				lineBytes := []byte(line[7:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if verbose {
						color.HiYellow("* #genre wasn't able to be converted (Line: %d)", lineIndex)
					}
				}
				convertedFile.Metadata.Tags = b
			} else if strings.HasPrefix(lineLower, "#maker") {
				if len(line) < 8 {
					if verbose {
						color.HiYellow("* #maker is invalid, ignoring (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[7:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if verbose {
						color.HiYellow("* #maker wasn't able to be converted (Line: %d)", lineIndex)
					}
				}
				convertedFile.Metadata.Creator = b
			} else if strings.HasPrefix(lineLower, "#title") {
				if len(line) < 8 {
					if verbose {
						color.HiYellow("* #title is invalid, replacing with default (Line: %d)", lineIndex)
					}
					convertedFile.Metadata.Title = "No title"
					continue
				}
				lineBytes := []byte(line[7:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if verbose {
						color.HiYellow("* #title wasn't able to be converted (Line: %d)", lineIndex)
					}
				}
				convertedFile.Metadata.Title = b
				if noBracketsInTitle {
					convertedFile.Metadata.Title = strings.Join(RegSplit(convertedFile.Metadata.Title, bracketRegex.String()), " ")
				}
			} else if strings.HasPrefix(lineLower, "#lnobj") {
				if len(line) < 8 {
					if verbose {
						color.HiRed("* #lnobj is invalid, cannot proceed (Line: %d)", lineIndex)
						return nil, nil
					}
					continue
				}
				if len(lineLower[7:]) != 2 {
					if verbose {
						color.HiRed("* #lnobj was specified, but not a valid hexatridecimal (Line: %d)", lineIndex)
						return nil, nil
					}
					continue
				}
				lnObject = lineLower[7:]
			} else if strings.HasPrefix(lineLower, "#artist") {
				if len(line) < 9 {
					if verbose {
						color.HiYellow("* #artist is invalid, but this will be ignored (Line: %d)", lineIndex)
					}
					continue
				}
				lineBytes := []byte(line[8:])
				b, e := BytesFromShiftJIS(lineBytes)
				if e != nil {
					if verbose {
						color.HiYellow("* #artist wasn't able to be converted (Line: %d)", lineIndex)
					}
				}
				convertedFile.Metadata.Artist = b
			} else if strings.HasPrefix(lineLower, "#playlevel") {
				if len(line) < 12 {
					if verbose {
						color.HiYellow("* #playlevel is invalid, but this will be ignored (Line: %d)", lineIndex)
					}
					continue
				}
				convertedFile.Metadata.Difficulty = line[11:]
			} else if strings.HasPrefix(lineLower, "#stagefile") {
				if len(line) < 12 {
					if verbose {
						color.HiYellow("* #stagefile is invalid, but this will be ignored (Line: %d)", lineIndex)
					}
					continue
				}
				convertedFile.Metadata.StageFile = line[11:]
			} else if strings.HasPrefix(lineLower, "#bpm ") {
				if len(line) < 6 {
					if verbose {
						color.HiYellow("* #bpm XX is invalid, parsing cannot continue (Line: %d)", lineIndex)
					}
					return nil, nil
				}
				i, e := strconv.ParseFloat(line[5:], 64)
				if e != nil {
					color.HiRed("* #bpm XX is not a number, parsing cannot continue (Line: %d)", lineIndex)
					return nil, nil
				}
				// The INITIAL BPM. BPM CHANGES are done through channel 2 or 8, depending on implementation.
				startTrackWithBPM = i
				convertedFile.TimingPoints[0.0] = startTrackWithBPM
			} else if strings.HasPrefix(lineLower, "#bpm") {
				// #BPMXX ....
				// Map BPM to corresponding hex value.
				if len(line) < 8 {
					color.HiYellow("* BPM change invalid, substituting with 0.0 (Line: %d)", lineIndex)
					bpmChanges[lineLower[4:6]] = 0.0
					continue
				}
				i, e := strconv.ParseFloat(line[7:], 64)
				if e != nil {
					// String
					color.HiYellow("* BPM change is not a number, substituting with 0.0 (Line: %d)", lineIndex)
					bpmChanges[lineLower[4:6]] = 0.0
					continue
				}
				bpmChanges[lineLower[4:6]] = i
			} else if strings.HasPrefix(lineLower, "#bmp") {
				// #BMPXX ....
				// Background animation.
				if len(line) < 8 {
					color.HiYellow("* BMP invalid, ignoring (Line: %d)", lineIndex)
					continue
				}
				bga[lineLower[4:6]] = line[7:]
			} else if strings.HasPrefix(lineLower, "#stop") {
				if len(line) < 9 {
					color.HiYellow("* STOP command isn't correctly formatted, not going to use it (Line: %d)", lineIndex)
					continue
				}
				// #STOPXX .....
				// Map STOP instructions to how long they need to be.
				i, e := strconv.ParseFloat(line[8:], 64)
				if e != nil {
					color.HiYellow("* STOP command is invalid, not going to use it (Line: %d)", lineIndex)
					continue
				}
				if i < 0.0 {
					color.HiYellow("* STOP command is negative (< 0.0), not going to use it (Line: %d)", lineIndex)
					continue
				}
				stopCommands[lineLower[5:7]] = i
			} else if strings.HasPrefix(lineLower, "#wav") {
				if len(line) < 8 {
					color.HiYellow("* WAV command invalid, all notes/sfx associated with it won't be placed (Line: %d)", lineIndex)
					continue
				}

				// Correct the extension used. E.g. if it's #WAV FILE.mp3 but actually FILE.wav, this will fix that.
				soundEffect := SearchForSoundFile(inputPath, line[7:])
				if len(soundEffect) == 0 {
					color.HiYellow("* %s isn't a file that could be found anywhere or isn't either .wav, .mp3, .ogg, or .3gp. (Line: %d)", line[7:], lineIndex)
					continue
				}
				convertedFile.KeySoundStringArray = append(convertedFile.KeySoundStringArray, soundEffect)
				hitSoundHexArray = append(hitSoundHexArray, lineLower[4:6])
			}
			lineIndex++
			continue
		}

		tInt, e := strconv.ParseInt(line[1:4], 10, 64)
		if e != nil {
			color.HiRed("* Failed to parse track #, cannot continue parsing (Line: %d, Content: %s)", lineIndex, line)
			return nil, nil
		}
		channel := lineLower[4:6]

		if len(mineRegex.FindString(channel)) > 0 {
			color.HiRed("* Cannot parse maps with mines (Line: %d)", lineIndex)
			return nil, nil
		}
		thisLineData := LocalTrackData{
			Channel: channel,
		}
		if len(line) > 7 {
			thisLineData.Message = strings.ToLower(line[7:])
		}
		TrackData[int(tInt)] = append(TrackData[int(tInt)], thisLineData)
		lineIndex++
	}
	// Didn't find #MAKER header
	if convertedFile.Metadata.Creator == "" {
		convertedFile.Metadata.Creator = convertedFile.Metadata.Artist
	}

	keys := make([]int, 0)
	for k := range TrackData {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		// Local values.
		scalar := 1.0
		// Ignore any duplicates that may exist.
		// Keep track of where exactly in the track time change directives happen.
		tempoChangeTimestampMap := make([]LocalTempoChange, 0)
		stopCommandTimestampMap := make([]LocalStopCommand, 0)

		// Read over all track data.
		for _, ltd := range TrackData[k] {
			switch ltd.Channel {
			case "02":
				i, e := strconv.ParseFloat(ltd.Message, 64)
				if e != nil {
					color.HiRed("* Failed to parse scalar, cannot continue parsing (Track: %d)", k)
					return nil, nil
				}
				scalar = i
				continue
			case "08", "03":
				if len(ltd.Message) == 0 {
					// #XXX08:
					// Unknown BPM if specified like this. Just going to assume it's 0.
					// And yeah, apparently this can happen
					tempoChangeTimestampMap = append(tempoChangeTimestampMap, LocalTempoChange{
						Position:   0.0,
						Bpm:        0.0,
						IsNegative: false,
					})
					continue
				}
				for i := 0; i < len(ltd.Message)/2; i++ {
					// Make no change for null values.
					if ltd.Message[i*2:(i*2)+2] == "00" {
						continue
					}
					var cBpm float64
					if ltd.Channel == "03" {
						// Channel 03 is used for hexadecimal BPM changes from 1-255.
						// As such, we need to parse it.
						parsedBpm, e := strconv.ParseInt(ltd.Message[i*2:(i*2)+2], 16, 64)
						if e != nil {
							cBpm = 0.0
						} else {
							cBpm = float64(parsedBpm)
						}
					} else {
						// Lookup BPM change in the map.
						cBpm = bpmChanges[ltd.Message[i*2:(i*2)+2]]
					}
					// At what % in the track does this BPM change happen?
					fraction := float64(i) / (float64(len(ltd.Message)) / 2.0)
					tempoChangeTimestampMap = append(tempoChangeTimestampMap, LocalTempoChange{
						Position:   fraction,
						Bpm:        math.Abs(cBpm),
						IsNegative: cBpm < 0.0,
					})
				}
				continue
			case "09":
				if len(ltd.Message) == 0 {
					// Blank STOP command. Don't process it.
					continue
				}
				for i := 0; i < len(ltd.Message)/2; i++ {
					if ltd.Message[i*2:(i*2)+2] == "00" {
						continue
					}
					// True if a STOP command for said message part appears in the known STOP command lengths.
					// if not, it was probably invalid.
					if val, ok := stopCommands[ltd.Message[i*2:(i*2)+2]]; ok {
						stopCommandTimestampMap = append(stopCommandTimestampMap, LocalStopCommand{
							Duration: val,
							Position: float64(i) / (float64(len(ltd.Message)) / 2.0),
						})
					}
				}
				continue
			}
		}

		// Put everything in order (if it isn't already). Avoid unnecessary sorts
		if len(tempoChangeTimestampMap) > 0 {
			sort.Slice(tempoChangeTimestampMap, func(i, j int) bool {
				return tempoChangeTimestampMap[i].Position < tempoChangeTimestampMap[j].Position
			})
		}
		if len(stopCommandTimestampMap) > 0 {
			sort.Slice(stopCommandTimestampMap, func(i, j int) bool {
				return stopCommandTimestampMap[i].Position < stopCommandTimestampMap[j].Position
			})
		}

		//stopCommandTimestampMap = RemoveOverlappingStopCommands(stopCommandTimestampMap)

		// Place objects down
		for _, ltd := range TrackData[k] {
			if len(ltd.Message)%2 != 0 {
				continue
			}
			if player2NoteRegex.MatchString(ltd.Channel) {
				if verbose {
					color.HiYellow("* This map has notes in player 2's side, which would overlap player 1. Not going to process this map.")
					return nil, nil
				}
			}
			if !(noteRegex.MatchString(ltd.Channel) || lnRegex.MatchString(ltd.Channel) || ltd.Channel == "01" || ltd.Channel == "04" || ltd.Channel == "07") {
				continue
			}
			for i := 0; i < len(ltd.Message)/2; i++ {
				if (i*2)+2 > len(ltd.Message) {
					break
				}
				target := ltd.Message[i*2 : (i*2)+2]
				// null
				if target == "00" {
					continue
				}
				localOffset := GetOffsetFromStartingTime(startTrackWithBPM, tempoChangeTimestampMap, stopCommandTimestampMap, i, ltd.Message, scalar)
				sfx := GetCorrespondingHitSound(hitSoundHexArray, target, volume)
				if noteRegex.MatchString(ltd.Channel) || lnRegex.MatchString(ltd.Channel) {
					laneInt := strings.Index(Base36Range, ltd.Channel[1:])
					hitObject := HitObject{
						Lane:      laneInt,
						StartTime: currentTime + localOffset,
					}
					// Uses the channel for the scratch lane, manually adjust to lane 8
					if laneInt == 6 {
						hitObject.Lane = 8
					} else if laneInt >= 8 {
						hitObject.Lane -= 2
					}
					// This file is for 9K, 14K, etc. so it is not compatible.
					if hitObject.Lane > 8 {
						color.HiRed("* File wants more than 8 keys, skipping")
						return nil, nil
					}

					// Closes the long note
					if len(lnObject) > 0 && target == lnObject {
						if len(convertedFile.HitObjects) == 0 {
							// Why is there an LN tail as the first object??
							continue
						}
						back := len(convertedFile.HitObjects) - 1
						if convertedFile.HitObjects[back].KeySounds == nil {
							// Previous hit object didn't have key sounds.
							// That means the previous value is a LNOBJ, so we ignore THE CURRENT LNOBJ.
							continue
						}
						if convertedFile.HitObjects[back].StartTime >= hitObject.StartTime {
							continue
						}
						convertedFile.HitObjects[back].IsLongNote = true
						convertedFile.HitObjects[back].EndTime = hitObject.StartTime
						continue
					}
					if sfx != nil {
						hitObject.KeySounds = sfx
					}

					// This is a long note existing in channels 51-59. We save it to a map storing these values.
					if lnRegex.MatchString(ltd.Channel) {
						// This is the end of a long note. Now, we can place the note.
						if longNoteTracker[hitObject.Lane] != 0.0 {
							// haha funny end time joke
							hitObject.EndTime = hitObject.StartTime
							hitObject.StartTime = longNoteTracker[hitObject.Lane]
							hitObject.IsLongNote = true
							if longNoteSoundEffectTracker[hitObject.Lane] != nil {
								hitObject.KeySounds = &HitObjectKeySound{
									Sample: longNoteSoundEffectTracker[hitObject.Lane].Sample,
									Volume: longNoteSoundEffectTracker[hitObject.Lane].Volume,
								}
							}
							// Reset values
							longNoteTracker[hitObject.Lane] = 0.0
							longNoteSoundEffectTracker[hitObject.Lane] = nil
							// Invalid long note because it ends either before or exactly at the position it ends.
							// In other words, do not process it.
							if hitObject.EndTime <= hitObject.StartTime {
								continue
							}
						} else {
							// This is the head of a long note, so we store its start time and key sounds for later.
							longNoteTracker[hitObject.Lane] = hitObject.StartTime
							longNoteSoundEffectTracker[hitObject.Lane] = hitObject.KeySounds
							continue
						}
					}
					convertedFile.HitObjects = append(convertedFile.HitObjects, hitObject)
					continue
				}
				if ltd.Channel == "01" {
					// Sound effect (channel 01) OR invisible note (which plays a sound effect).
					soundEffect := SoundEffect{
						StartTime: currentTime + localOffset,
					}
					if sfx != nil {
						soundEffect.Sample = sfx.Sample
						soundEffect.Volume = sfx.Volume
					} else {
						// No sound effect assigned to this address?
						continue
					}
					convertedFile.SoundEffects = append(convertedFile.SoundEffects, soundEffect)
				}
				if ltd.Channel == "04" || ltd.Channel == "07" {
					t := bga[target]
					l := Back
					if ltd.Channel == "07" {
						l = Front
					}
					if len(t) > 0 {
						convertedFile.BackgroundAnimation = append(convertedFile.BackgroundAnimation, BackgroundAnimation{
							StartTime: currentTime + localOffset,
							File:      t,
							Layer:     l,
						})
					}
				}
			}
		}

		// Get the full length of the track, factoring STOP commands.
		fullLengthOfTrack := GetTotalTrackDuration(startTrackWithBPM, tempoChangeTimestampMap, stopCommandTimestampMap, scalar) // Calculate all timing points.
		timingPoints := CalculateTimingPoints(currentTime, startTrackWithBPM, tempoChangeTimestampMap, stopCommandTimestampMap, scalar)
		//timingPoints = RemoveOverlappingTimingPoints(timingPoints)
		for k, v := range timingPoints {
			convertedFile.TimingPoints[k] = v
		}

		if len(tempoChangeTimestampMap) > 0 {
			startTrackWithBPM = tempoChangeTimestampMap[len(tempoChangeTimestampMap)-1].Bpm
		}

		// Adjust current time.
		currentTime += fullLengthOfTrack
		convertedFile.TimingPoints[currentTime] = startTrackWithBPM
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	sort.Slice(convertedFile.BackgroundAnimation, func(i, j int) bool {
		return convertedFile.BackgroundAnimation[i].StartTime < convertedFile.BackgroundAnimation[j].StartTime
	})

	return &convertedFile, nil
}
