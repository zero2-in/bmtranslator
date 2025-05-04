package main

const (
	// MinuteUnit is the unit for one (1) minute, in seconds.
	MinuteUnit = 60.0

	// Second is the conversion of one thousand (1,000) milliseconds to one second.
	Second = 1000.0

	// DefaultStartingBPM must be 130, according to the original BM98 format specification.
	// http://bm98.yaneu.com/bm98/bmsformat.html
	DefaultStartingBPM = 130.0

	// TempDir is the directory name used by BMT to temporarily store files.
	// It is removed when the program exits, regardless of status.
	TempDir = "bmt_temp_folder"

	// Version is the current version of the program.
	Version = "0.2.2"

	// JSONVersion is the current version of the JSON output of the program.
	// Also included in JSON files as a "version" key.
	JSONVersion = "v1"

	// Base36Range is used for lane conversion.
	Base36Range = "0123456789abcdefghijklmnopqrstuvwxyz"
)
