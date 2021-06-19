package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func RecursiveMultiPathZip(path1, path2, destinationPath string) error {
	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	err = walkPath(path1, myZip)
	if err != nil {
		return err
	}
	err = walkPath(path2, myZip)
	if err != nil {
		return err
	}
	err = myZip.Close()
	if err != nil {
		return err
	}
	return nil
}

func walkPath(path string, z *zip.Writer) error {
	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filepath.FromSlash(filePath), filepath.FromSlash(path)+string(os.PathSeparator))
		zipFile, err := z.Create(relPath)
		if err != nil {
			return err
		}
		fsFile, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer fsFile.Close()
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			return err
		}
		return nil
	})
}
