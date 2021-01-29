package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/fatih/color"
	copy2 "github.com/otiai10/copy"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type ConversionStatus struct {
	Name    string
	Success int
	Fail    int
	Skip    bool
}

func main() {
	fmt.Print("\n")
	color.HiBlue("  ___ __  __ _____ ")
	color.HiBlue(" | _ )  \\/  |_   _|")
	color.HiBlue(" | _ \\ |\\/| | | |  ")
	color.HiBlue(" |___/_|  |_| |_|  (version %s)", Version)
	fmt.Print("\n\n")

	conf := NewProgramConfig()

	if conf.Input == "-" {
		log.Fatal("Input directory must be provided. Use the argument -i /path/to/input to define a directory with BMS folders.")
	}
	if conf.Output == "-" {
		log.Fatal("Output directory must be provided. Use the argument -o /path/to/output to define where BMT will output files.")
	}

	if conf.Verbose {
		color.HiBlack("* Input wanted: %s", conf.Input)
		color.HiBlack("* Output wanted: %s", conf.Output)
		color.HiBlack("* Volume: %d", conf.Volume)
		color.HiBlack("* Verbose: %t", conf.Verbose)
		color.HiBlack("* Dump file data to output: %t", conf.DumpFileData)
	}

	if conf.FileType == Osu {
		color.HiBlue("* osu! specific: Using OD %f and HP %f.", conf.OverallDifficulty, conf.HPDrain)
	}
	// Check existence of output directory
	_, err := os.Stat(filepath.FromSlash(conf.Output))
	if err != nil {
		log.Fatal(err)
	}

	// Check if temp folder is still left over
	_, err = os.Stat(path.Join(filepath.FromSlash(conf.Output), TempDir))
	if err != nil {
		if os.IsNotExist(err) {
			if !strings.Contains(err.Error(), "The system cannot find the file specified.") {
				color.HiRed("* Failed to check for existence of the temporary directory. error: %s", err.Error())
				return
			} else {
				if conf.Verbose {
					color.HiBlack("* Temporary directory does not exist yet.")
				}
			}
		}
	} else {
		if conf.Verbose {
			color.HiBlack("* Temporary directory appears to already exist. Attempting to remove")
		}
		err = os.RemoveAll(path.Join(filepath.FromSlash(conf.Output), TempDir))
		if err != nil {
			color.HiRed("* Failed to remove temporary directory. Location: %s error: %s", path.Join(filepath.FromSlash(conf.Output), TempDir), err.Error())
			return
		}
		if conf.Verbose {
			color.HiBlack("* Deleted old temporary directory successfully")
		}
	}

	if conf.Verbose {
		color.HiBlack("* Decoding background image to memory.")
	}

	var im image.Image
	idx := strings.Index(BgImage, ";base64,")
	decoded, err := base64.StdEncoding.DecodeString(BgImage[idx+8:])
	if err != nil {
		color.HiRed("* Failed to parse background image base64. %s", err.Error())
	} else {
		r := bytes.NewReader(decoded)
		im, _ = png.Decode(r)
	}

	// Scan input directory
	inputFolders, err := ioutil.ReadDir(filepath.FromSlash(conf.Input))
	if err != nil {
		log.Fatal(err)
	}
	if len(inputFolders) == 0 {
		color.HiRed("* No folders found in input directory.")
		return
	}

	if conf.Verbose {
		color.HiBlack("* Found %d directories to process.", len(inputFolders))
	}

	// Create temporary directory
	err = os.Mkdir(path.Join(filepath.FromSlash(conf.Output), TempDir), 0700)
	if err != nil {
		color.HiRed("* Could not create a temporary folder inside the output directory. %s", err.Error())
		return
	}

	// Store statuses
	var conversionStatus []ConversionStatus

	// Iterate over all files
	for fI, f := range inputFolders {
		color.White("* [%d/%d] Processing %s", fI+1, len(inputFolders), color.YellowString(f.Name()))
		conversionStatus = append(conversionStatus, ConversionStatus{})
		conversionStatus[fI].Name = f.Name()
		if !f.IsDir() {
			color.HiRed("* %s is not a directory. Skipping.", f.Name())
			conversionStatus[fI].Skip = true
			continue
		}
		input := filepath.ToSlash(path.Join(filepath.FromSlash(conf.Input), f.Name()))
		output := filepath.ToSlash(path.Join(filepath.FromSlash(conf.Output), TempDir, f.Name()))

		var bmsChartFiles []string
		files, err := ioutil.ReadDir(input)
		if err != nil {
			conversionStatus[fI].Skip = true
			color.HiRed("* Failed to read directory of %s. Skipping. (Error: %s)", f.Name(), err.Error())
			continue
		}
		// Most BMS zip files appear to be nested :^(
		if len(files) == 1 && files[0].IsDir() {
			input = filepath.ToSlash(path.Join(filepath.FromSlash(conf.Input), f.Name(), files[0].Name()))
			files, err = ioutil.ReadDir(input)
			if err != nil {
				conversionStatus[fI].Skip = true
				color.HiRed("* Failed to read directory of %s. Skipping. (Error: %s)", f.Name(), err.Error())
				continue
			}
		}

		if len(files) == 0 {
			conversionStatus[fI].Skip = true
			color.HiRed("* No files are in %s. Skipping.", f.Name())
			continue
		}

		// Iterate over all files
		for _, f := range files {
			if f.Size() == 0 || f.IsDir() {
				continue
			}
			if strings.HasSuffix(f.Name(), ".bms") || strings.HasSuffix(f.Name(), ".bme") || strings.HasSuffix(f.Name(), ".bml") {
				bmsChartFiles = append(bmsChartFiles, f.Name())
			}
		}
		if len(bmsChartFiles) == 0 {
			conversionStatus[fI].Skip = true
			color.HiRed("* Didn't find any .bms, .bme or .bml files in %s. Skipping.", f.Name())
			continue
		}

		// Copy all contents
		if conf.Verbose {
			color.HiBlack("* Checks passed; copying %s to %s.", input, output)
		}
		err = copy2.Copy(input, output)
		if err != nil {
			conversionStatus[fI].Skip = true
			color.HiRed("* Failed to copy %s. Skipping. (%s)", f.Name(), err.Error())
			continue
		}

		if conf.Verbose {
			color.HiBlack("* Copy succeeded, found %d charts.", len(bmsChartFiles))
		}

		zipExtension := "qp"
		fileExtension := "qua"
		switch conf.FileType {
		case Osu:
			zipExtension = "osz"
			fileExtension = "osu"
		}
		for diffIndex, bmsFile := range bmsChartFiles {
			if conf.Verbose {
				color.HiBlack("* [%d/%d] Converting %s -> %s file type", diffIndex+1, len(bmsChartFiles), bmsFile, fileExtension)
			}
			fileData, err := conf.GetFileData(output, bmsFile)
			if err != nil {
				conversionStatus[fI].Fail++
				color.HiRed("* %s wasn't parsed due to an error: %s", bmsFile, err.Error())
				continue
			}
			if fileData == nil {
				color.HiYellow("* %s was skipped", bmsFile)
				conversionStatus[fI].Fail++
				continue
			}

			if conf.Verbose {
				color.HiBlack("* Processed %d hit objects, %d sound effects and %d timing points", len(fileData.HitObjects), len(fileData.SoundEffects), len(fileData.TimingPoints))
				if conf.FileType == Osu {
					color.HiBlack("* osu! specific: found %d background animation frames", len(fileData.BackgroundAnimation))
				}
			}

			if len(fileData.Meta.StageFile) == 0 {
				// Copy BG image
				bgFile, err := os.OpenFile(path.Join(filepath.FromSlash(conf.Output), TempDir, f.Name(), "bg.png"), os.O_WRONLY|os.O_CREATE, 0777)
				if err != nil {
					color.HiRed("* Failed to open a new file for the background; ignoring (%s)", err.Error())
				} else {
					e := png.Encode(bgFile, im)
					if e != nil {
						color.HiRed("* Failed to encode the background; ignoring (%s)", e.Error())
					}
					fileData.Meta.StageFile = "bg.png"
					bgFile.Close()
				}
			}

			writeTo := path.Join(output, strings.TrimSuffix(bmsFile, path.Ext(bmsFile))+"."+fileExtension)
			switch conf.FileType {
			case Osu:
				err = conf.ConvertBmsToOsu(*fileData, writeTo)
				break
			default:
				err = ConvertBmsToQua(*fileData, writeTo)
			}
			if err != nil {
				conversionStatus[fI].Fail++
				color.HiYellow("* %s wasn't written to due to an error: %s", bmsFile, err.Error())
				continue
			}

			if conf.DumpFileData {
				bmsFileName := strings.TrimSuffix(bmsFile, path.Ext(bmsFile))
				err = conf.WriteDump(*fileData, path.Join(filepath.FromSlash(conf.Output), f.Name()+"-"+bmsFileName+".txt"), bmsFileName)
				if err != nil && conf.Verbose {
					color.HiYellow("* failed to write dump for %s: %s (conversion still succeeded)", bmsFile, err.Error())
				}
			}

			conversionStatus[fI].Success++
		}

		if err := RecursiveZip(output, path.Join(filepath.FromSlash(conf.Output), f.Name()+"."+zipExtension)); err != nil {
			panic(err)
		}

		if conf.Verbose {
			color.HiBlack("* ---- Done ----")
		}
	}

	color.HiGreen("* Finished conversion of all queued folders. Find them in %s", conf.Output)
	for _, s := range conversionStatus {
		if s.Skip {
			color.HiYellow("* %s was skipped", s.Name)
			continue
		}
		color.White("* %s: %d %s and %d %s", s.Name, s.Fail, color.YellowString("not converted"), s.Success, color.HiGreenString("succeeded"))
	}

	e := os.RemoveAll(path.Join(filepath.FromSlash(conf.Output), TempDir))
	if e != nil {
		color.HiRed("* Failed to remove temp dir. %s", e.Error())
	} else {
		if conf.Verbose {
			color.HiBlack("* Cleaned up temp dir successfully.")
		}
	}
}
