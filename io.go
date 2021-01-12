package main

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

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
