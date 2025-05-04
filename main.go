package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

func welcomeToOss() {
	fmt.Print("\n")
	str := `██████╗ ███╗   ███╗████████╗
██╔══██╗████╗ ████║╚══██╔══╝
██████╔╝██╔████╔██║   ██║   
██╔══██╗██║╚██╔╝██║   ██║   
██████╔╝██║ ╚═╝ ██║   ██║   
╚═════╝ ╚═╝     ╚═╝   ╚═╝`
	scanner := bufio.NewScanner(strings.NewReader(str))
	for scanner.Scan() {
		color.HiBlue(scanner.Text())
	}
	color.HiBlue("Version %s", Version)
	fmt.Print("\n\n")
}

func main() {

	welcomeToOss()
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
		color.HiBlack("* Additional JSON output: %t", conf.JSONOutput)
	}

	if conf.FileType == Osu {
		color.HiBlue("* osu! specific: Using OD %f and HP %f.", conf.OverallDifficulty, conf.HPDrain)
	}

	if conf.JSONOnly {
		color.HiYellow("* JSON only is enabled: Output zips to import into games won't be created; only JSON files will be created.")
	}
	// Check existence of input & output directory
	inputExists := FileExists(conf.Input)
	if !inputExists {
		color.HiRed("* Input directory does not exist.")
		return
	}

	outputExists := FileExists(conf.Output)
	if !outputExists {
		color.HiRed("* Output directory does not exist.")
		return
	}

	// Check if temp folder is still left over
	_, err := os.Stat(path.Join(filepath.FromSlash(conf.Output), TempDir))
	if err != nil {
		if os.IsNotExist(err) {
			if conf.Verbose {
				color.HiBlack("* Temporary directory does not exist yet.")
			}
		} else {
			color.HiRed("* Failed to check for existence of the temporary directory. error: %s", err.Error())
			return
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
			// include .BME, .Bme, .bMe, .bmE ...
			lower := strings.ToLower(f.Name())
			if strings.HasSuffix(lower, ".bms") || strings.HasSuffix(lower, ".bml") || strings.HasSuffix(lower, ".bme") {
				bmsChartFiles = append(bmsChartFiles, f.Name())
			}
		}
		if len(bmsChartFiles) == 0 {
			conversionStatus[fI].Skip = true
			color.HiRed("* Didn't find any .bms, .bme or .bml files in %s. Skipping.", f.Name())
			continue
		}

		err = os.Mkdir(output, 0755)
		if err != nil {
			color.HiRed("* Failed to create a folder for %s. Skipping. (%s)", f.Name(), err.Error())
			continue
		}
		// Copy all contents
		//if conf.Verbose {
		//	color.HiBlack("* Checks passed; copying %s to %s.", input, output)
		//}
		//err = copy2.Copy(input, output)
		//if err != nil {
		//	conversionStatus[fI].Skip = true
		//	color.HiRed("* Failed to copy %s. Skipping. (%s)", f.Name(), err.Error())
		//	continue
		//}

		if conf.Verbose {
			color.HiBlack("* Found %d charts to process", len(bmsChartFiles))
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
				if conf.JSONOnly {
					color.HiBlack("* [%d/%d] %s -> .json", diffIndex+1, len(bmsChartFiles), bmsFile)
				} else {
					color.HiBlack("* [%d/%d] %s -> .%s ", diffIndex+1, len(bmsChartFiles), bmsFile, fileExtension)
				}
			}
			fileData, err := conf.ReadFileData(input, bmsFile)
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

			if conf.FileType == Osu && conf.Verbose {
				color.HiBlack("* osu! specific: found %d background animation frames", len(fileData.BGAFrames))
			}

			if conf.JSONOutput || conf.JSONOnly {
				bmsFileName := strings.TrimSuffix(bmsFile, path.Ext(bmsFile))
				err = conf.ConvertBmsToJson(*fileData, path.Join(conf.Output, f.Name()+" - "+bmsFileName+".json"))
				if err != nil && conf.Verbose {
					color.HiRed("* failed to write json for %s: %s", bmsFile, err.Error())
				}
			}

			if !conf.JSONOnly {
				writeTo := path.Join(output, strings.TrimSuffix(bmsFile, path.Ext(bmsFile))+"."+fileExtension)
				switch conf.FileType {
				case Osu:
					err = conf.ConvertBmsToOsu(*fileData, writeTo)
					break
				default:
					err = conf.ConvertBmsToQua(*fileData, writeTo)
				}
				if err != nil {
					conversionStatus[fI].Fail++
					color.HiYellow("* %s wasn't written to due to an error: %s", bmsFile, err.Error())
					continue
				}
			}

			conversionStatus[fI].Success++
		}

		if !conf.JSONOnly {
			if err := RecursiveMultiPathZip(input, output, path.Join(conf.Output, f.Name()+"."+zipExtension)); err != nil {
				panic(err)
			}
		}

		if conf.Verbose {
			color.HiBlack("* ---- Done with this folder ----")
		}
	}

	color.HiGreen("* Finished conversion of all queued folders.")
	for _, s := range conversionStatus {
		if s.Skip {
			color.HiYellow("* %s was skipped", s.Name)
			continue
		}
		color.White("* %s: %d %s and %d %s", s.Name, s.Fail, color.YellowString("not converted"), s.Success, color.HiGreenString("succeeded"))
	}

	e := os.RemoveAll(path.Join(conf.Output, TempDir))
	if e != nil {
		color.HiRed("* Failed to remove temp dir. %s", e.Error())
	} else {
		if conf.Verbose {
			color.HiBlack("* Cleaned up temp dir successfully.")
		}
	}
}
