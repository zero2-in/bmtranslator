package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// BMS #WAV values can have mismatching extensions (sometimes it's .wav when it's actually .ogg on the filesystem).
// This will correct the extension, or return nothing if it wasn't found.
func SearchForSoundFile(originalPath string, pathToSoundFile string) string {
	possibleExtensions := []string{".wav", ".mp3", ".ogg", ".3gp"}

	originalPath = filepath.ToSlash(originalPath)
	pathToSoundFileNoExt := filepath.ToSlash(strings.TrimSuffix(pathToSoundFile, path.Ext(pathToSoundFile)))
	for _, extension := range possibleExtensions {
		_, e := os.Stat(path.Clean(path.Join(originalPath, pathToSoundFileNoExt) + extension))
		if e != nil {
			continue
		}
		return pathToSoundFileNoExt + extension
	}

	return ""
}

// Call os.Stat to see if a file exists or not. Returns true if it does.
func FileExists(location string) bool {
	_, e := os.Stat(location)
	return e == nil
}
