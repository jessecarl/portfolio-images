// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		imageSizes   imageSizeSlice
		inputGlob    string
		outputDir    string
		forceSave    bool
		imageQuality int
	)
	flag.Var(&imageSizes, "s", "[required] Image Sizes to be output. Should be of the form \"suffix:size\" where \"suffix\" is added to the base filename when saving this version, and \"size\" defines the maximum square dimension (0 for no resize).")
	flag.StringVar(&inputGlob, "i", "", "[required] Glob used to find all images to be resized.")
	flag.StringVar(&outputDir, "o", "", "[required] Output directory to save all generated images to.")
	flag.BoolVar(&forceSave, "f", false, "Force overwrite of existing images in the output directory. Default is false.")
	flag.IntVar(&imageQuality, "q", 80, "Image quality 1-100 inclusive.")
	flag.Parse()

	if len(imageSizes) == 0 || len(inputGlob) == 0 || len(outputDir) == 0 || imageQuality < 1 || imageQuality > 100 {
		flag.PrintDefaults()
		os.Exit(2)
	}

	if err := createOutputDirIfNotExist(outputDir); err != nil {
		log.Fatal(err)
	}

	// find all input files
	inputFiles, err := filepath.Glob(inputGlob)
	if err != nil {
		log.Fatalf("Error parsing input glob, %q: %v", inputGlob, err)
	} else if len(inputFiles) == 0 {
		log.Fatalf("No files found for input glob, %q", inputGlob)
	}

	collectedErrors := []error{}
	for _, inputFile := range inputFiles {
		// Open Valid Input Image
		ii, err := NewImageInput(inputFile)
		if err != nil {
			collectedErrors = append(collectedErrors, err)
			continue
		}

		for _, size := range imageSizes {
			// Open Valid Output File
			out, err := NewImageOutput(ii, size, outputDir, forceSave)
			if os.IsExist(err) {
				log.Printf("Skipping %q %v, as image already exists", ii.Filename, size)
				continue
			} else if err != nil {
				collectedErrors = append(collectedErrors, err)
				continue
			}

			// Resize to Fit
			out.Transform()

			// Save to file
			err = out.Save(imageQuality)
			if err != nil {
				collectedErrors = append(collectedErrors, err)
				continue
			}
		}
	}

	if len(collectedErrors) > 0 {
		log.Printf("Errors during processing: %v", collectedErrors)
	}
}

func createOutputDirIfNotExist(outputDir string) error {
	info, err := os.Stat(outputDir)
	// determine if the directory exists and is a directory
	if err == nil {
		if info.IsDir() {
			return nil
		} else {
			return fmt.Errorf("Cannot replace file with directory, %q", outputDir)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Error Checking Output Directory, %q: %v", outputDir, err)
	}
	err = os.MkdirAll(outputDir, os.ModePerm|os.ModeDir)
	if err != nil {
		return fmt.Errorf("Error Creating Output Directory, %q: %v", outputDir, err)
	}
	return nil
}

/*

Proposed Pipeline:

Queue:InputFiles --(filename)->
-> Filter:OpenValidInputFiles --(image,filename)->
-> Multiplexer:AllSizesForImage(imageSizes) --(image,size,filename)->
-> Filter:OpenValidOutputFiles(force) --(image,size,outfile)->
-> Worker:ResizeToFit --(image,outfile)->
-> Worker:Save

*/
