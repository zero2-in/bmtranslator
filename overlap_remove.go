package main

// Same as RemoveOverlappingTimingPoints. However, Go doesn't have any generics, so it must be copied...
func RemoveOverlappingStopCommands(s []LocalStopCommand) []LocalStopCommand {
	timingPointMap := make(map[float64]float64)
	for _, s2 := range s {
		timingPointMap[s2.Position] = s2.Duration
	}
	arr := make([]LocalStopCommand, len(timingPointMap))
	for k, v := range timingPointMap {
		arr = append(arr, LocalStopCommand{
			Position: k,
			Duration: v,
		})
	}
	return arr
}

// Timing points created by CalculateTimingPoints can overlap, somewhat. This function will try to remove them.
//func RemoveOverlappingTimingPoints(s []TimingPoint) []TimingPoint {
//	timingPointMap := make(map[float64]float64)
//	for _, s2 := range s {
//		timingPointMap[s2.StartTime] = s2.Bpm
//	}
//	arr := make([]TimingPoint, len(timingPointMap))
//	for k, v := range timingPointMap {
//		arr = append(arr, TimingPoint{
//			StartTime: k,
//			Bpm: v,
//		})
//	}
//	return arr
//}
