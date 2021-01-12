package main

import (
	"bytes"
	"encoding/base64"
	"flag"
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
	color.HiBlue("...---=== BMTranslator %s ====---...", Version)
	fmt.Print("\n")

	inputPtr := flag.String("i", "-", "Input folder containing BMS FOLDERS, NOT zip files!")
	outputPtr := flag.String("o", "-", "Which folder you want the ending .qua files to be output to")
	volPtr := flag.Int("vol", 100, "How loud the key sounds should be (0-100 is acceptable)")
	fileTypePtr := flag.String("type", "quaver", "Which file type to use. (quaver | osu)")
	hpDrainPtr := flag.Float64("hp", 8.5, "If file type is 'osu', the HP drain (0-10)")
	overallDifficultyPtr := flag.Float64("od", 8.0, "If file type is 'osu', the overall difficulty (0-10)")
	verbosePtr := flag.Bool("v", false, "If true, all logs will be shown.")
	noBracketsPtr := flag.Bool("no-brackets", false, "If true, all brackets and their contents ([like this]) will be removed from the title of the map.")

	// TODO: Implement 5K+1 alignment feature someday
	//specialNotePosPtr := flag.String("pos", "-", "If the style is 5K+1, where should the notes be aligned to? (left for 1-5, right for 3-7. Default is right.)")
	flag.Parse()

	if *inputPtr == "-" {
		log.Fatal("Input directory must be provided")
	}
	if *outputPtr == "-" {
		log.Fatal("Output directory must be provided")
	}

	if *verbosePtr {
		color.HiBlack("* Input wanted: %s", *inputPtr)
		color.HiBlack("* Output wanted: %s", *outputPtr)
		color.HiBlack("* Volume: %d", *volPtr)
		color.HiBlack("* Verbose: %t", *verbosePtr)
	}

	color.HiBlue("* Using file type \"%s\"", *fileTypePtr)

	//if *specialNotePosPtr == "-" {
	//	*specialNotePosPtr = "right"
	//}
	volInt := ClampInt(*volPtr, 100, 0)
	hpDrain := ClampFloat(*hpDrainPtr, 10.0, 0.0)
	overallDifficulty := ClampFloat(*overallDifficultyPtr, 10.0, 0.0)

	if *fileTypePtr == "osu" {
		color.HiBlue("* osu! specific: Using OD %f and HP %f.", overallDifficulty, hpDrain)
	}
	// Check existence of output directory
	_, err := os.Stat(filepath.FromSlash(*outputPtr))
	if err != nil {
		log.Fatal(err)
	}

	// Check if temp folder is still left over
	_, err = os.Stat(path.Join(filepath.FromSlash(*outputPtr), TempDir))
	if err != nil {
		if os.IsNotExist(err) {
			if !strings.Contains(err.Error(), "The system cannot find the file specified.") {
				color.HiRed("* Failed to check for existence of the temporary directory. error: %s", err.Error())
				return
			} else {
				if *verbosePtr {
					color.HiBlack("* Temporary directory does not exist yet.")
				}
			}
		}
	} else {
		if *verbosePtr {
			color.HiBlack("* Temporary directory appears to already exist. Attempting to remove")
		}
		err = os.RemoveAll(path.Join(filepath.FromSlash(*outputPtr), TempDir))
		if err != nil {
			color.HiRed("* Failed to remove temporary directory. Location: %s error: %s", path.Join(filepath.FromSlash(*outputPtr), TempDir), err.Error())
			return
		}
		if *verbosePtr {
			color.HiBlack("* Deleted old temporary directory successfully")
		}
	}

	if *verbosePtr {
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
	inputFolders, err := ioutil.ReadDir(filepath.FromSlash(*inputPtr))
	if err != nil {
		log.Fatal(err)
	}
	if len(inputFolders) == 0 {
		color.HiRed("* No folders found in input directory.")
		return
	}

	if *verbosePtr {
		color.HiBlack("* Found %d directories to process.", len(inputFolders))
	}

	// Create temporary directory
	err = os.Mkdir(path.Join(filepath.FromSlash(*outputPtr), TempDir), 0700)
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
		input := filepath.ToSlash(path.Join(filepath.FromSlash(*inputPtr), f.Name()))
		output := filepath.ToSlash(path.Join(filepath.FromSlash(*outputPtr), TempDir, f.Name()))

		var bmsChartFiles []string
		files, err := ioutil.ReadDir(input)
		if err != nil {
			conversionStatus[fI].Skip = true
			color.HiRed("* Failed to read directory of %s. Skipping. (Error: %s)", f.Name(), err.Error())
			continue
		}
		// Most BMS zip files appear to be nested :^(
		if len(files) == 1 && files[0].IsDir() {
			input = filepath.ToSlash(path.Join(filepath.FromSlash(*inputPtr), f.Name(), files[0].Name()))
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
		if *verbosePtr {
			color.HiBlack("* Checks passed; copying %s to %s.", input, output)
		}
		err = copy2.Copy(input, output)
		if err != nil {
			conversionStatus[fI].Skip = true
			color.HiRed("* Failed to copy %s. Skipping. (%s)", f.Name(), err.Error())
			continue
		}

		if *verbosePtr {
			color.HiBlack("* Copy succeeded, found %d charts.", len(bmsChartFiles))
		}

		zipExtension := "qp"
		fileExtension := "qua"
		switch *fileTypePtr {
		case "osu":
			zipExtension = "osz"
			fileExtension = "osu"
		}
		for diffIndex, bmsFile := range bmsChartFiles {
			if *verbosePtr {
				color.HiBlack("* [%d/%d] Converting %s -> %s file type", diffIndex+1, len(bmsChartFiles), bmsFile, fileExtension)
			}
			convertedFile, err := GetConvertedFile(output, bmsFile, *verbosePtr, volInt, *noBracketsPtr)
			if err != nil {
				conversionStatus[fI].Fail++
				color.HiRed("* %s wasn't parsed due to an error: %s", bmsFile, err.Error())
				continue
			}
			if convertedFile == nil {
				color.HiYellow("* %s was skipped", bmsFile)
				conversionStatus[fI].Fail++
				continue
			}

			if *verbosePtr {
				color.HiBlack("* Processed %d hit objects, %d sound effects and %d timing points", len(convertedFile.HitObjects), len(convertedFile.SoundEffects), len(convertedFile.TimingPoints))
			}

			if len(convertedFile.Metadata.StageFile) == 0 {
				// Copy BG image
				bgFile, err := os.OpenFile(path.Join(filepath.FromSlash(*outputPtr), TempDir, f.Name(), "bg.png"), os.O_WRONLY|os.O_CREATE, 0777)
				if err != nil {
					color.HiRed("* Failed to open a new file for the background. Ignoring the error. (%s)", err.Error())
				} else {
					e := png.Encode(bgFile, im)
					if e != nil {
						color.HiRed("* Failed to encode the background. Ignoring the error. (%s)", e.Error())
					}
					bgFile.Close()
				}
				convertedFile.Metadata.StageFile = "bg.png"
			}

			switch *fileTypePtr {
			case "osu":
				err = ConvertBmsToOsu(*convertedFile, path.Join(output, strings.TrimSuffix(bmsFile, path.Ext(bmsFile))+"."+fileExtension), hpDrain, overallDifficulty, volInt)
				break
			default:
				err = ConvertBmsToQua(*convertedFile, path.Join(output, strings.TrimSuffix(bmsFile, path.Ext(bmsFile))+"."+fileExtension))
			}
			if err != nil {
				conversionStatus[fI].Fail++
				color.HiYellow("* %s wasn't written to due to an error: %s", bmsFile, err.Error())
				continue
			}

			conversionStatus[fI].Success++
		}

		if err := RecursiveZip(output, path.Join(filepath.FromSlash(*outputPtr), f.Name()+"."+zipExtension)); err != nil {
			panic(err)
		}

		if *verbosePtr {
			color.HiBlack("* Done")
		}
	}

	color.HiGreen("* Done.")
	for _, s := range conversionStatus {
		if s.Skip {
			color.HiYellow("* %s was skipped", s.Name)
			continue
		}
		color.White("* %s: %d %s and %d %s", s.Name, s.Fail, color.YellowString("not converted"), s.Success, color.HiGreenString("succeeded"))
	}

	e := os.RemoveAll(path.Join(filepath.FromSlash(*outputPtr), TempDir))
	if e != nil {
		color.HiRed("* Failed to remove temp dir. %s", e.Error())
	} else {
		if *verbosePtr {
			color.HiBlack("* Cleaned up temp dir successfully.")
		}
	}
}
