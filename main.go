// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func main() {
	var (
		imageSizes   imageSizeSlice
		inputGlob    string
		outputDir    string
		forceSave    bool
		imageQuality int
		workerCount  int
	)
	flag.Var(&imageSizes, "s", "[required] Image Sizes to be output. Should be of the form \"suffix:size\" where \"suffix\" is added to the base filename when saving this version, and \"size\" defines the maximum square dimension (0 for no resize).")
	flag.StringVar(&inputGlob, "i", "", "[required] Glob used to find all images to be resized.")
	flag.StringVar(&outputDir, "o", "", "[required] Output directory to save all generated images to.")
	flag.BoolVar(&forceSave, "f", false, "Force overwrite of existing images in the output directory. Default is false.")
	flag.IntVar(&imageQuality, "q", 80, "Image quality 1-100 inclusive.")
	flag.IntVar(&workerCount, "w", 5, "Width of the pipeline or number of workers per pipeline stage.")
	flag.Parse()

	if len(imageSizes) == 0 || len(inputGlob) == 0 || len(outputDir) == 0 || imageQuality < 1 || imageQuality > 100 || workerCount < 1 {
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

	done := make(chan struct{})
	errc := make(chan error)
	go func(ec <-chan error) {
		for err := range ec {
			log.Printf("[WARNING]: %v", err)
		}
	}(errc)

	// Step One: Open Valid Input Image
	inputFileCh := QueueImages(done, inputFiles...)

	inputImageCh := make(chan *ImageInput)
	var inputWg sync.WaitGroup
	inputWg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			OpenImages(done, errc, inputFileCh, inputImageCh)
			inputWg.Done()
		}()
	}
	go func() {
		inputWg.Wait()
		close(inputImageCh)
	}()

	// Step Two: Open Valid Output File
	readyImageCh := make(chan *ImageOutput)
	var sizeWg sync.WaitGroup
	sizeWg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			ReadyImages(done, errc, inputImageCh, readyImageCh, outputDir, forceSave, imageSizes...)
			sizeWg.Done()
		}()
	}
	go func() {
		sizeWg.Wait()
		close(readyImageCh)
	}()

	// Step Three: Resize to Fit
	resizedImageCh := make(chan *ImageOutput)
	var resizeWg sync.WaitGroup
	resizeWg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			ResizeImages(done, readyImageCh, resizedImageCh)
			resizeWg.Done()
		}()
	}
	go func() {
		resizeWg.Wait()
		close(resizedImageCh)
	}()

	// Step Four: Save Image (end of pipeline)
	var savedWg sync.WaitGroup
	savedWg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			SaveImages(done, errc, resizedImageCh, imageQuality)
			savedWg.Done()
		}()
	}
	savedWg.Wait()
	close(done)
	close(errc)
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

func QueueImages(done <-chan struct{}, filenames ...string) <-chan string {
	out := make(chan string)
	go func() {
		defer close(out)
		for _, n := range filenames {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func OpenImages(done <-chan struct{}, errc chan<- error, filenames <-chan string, imgc chan<- *ImageInput) {
	for n := range filenames {
		var send chan<- *ImageInput
		var ec chan<- error

		ii, err := NewImageInput(n)
		if err != nil {
			ec = errc
		} else {
			send = imgc
		}

		select {
		case send <- ii:
		case ec <- err:
		case <-done:
			return
		}
	}
}

func ReadyImages(done <-chan struct{}, errc chan<- error, inc <-chan *ImageInput, outc chan<- *ImageOutput, outputDir string, force bool, sizes ...ImageSize) {
	for in := range inc {
		for _, size := range sizes {
			var send chan<- *ImageOutput
			var ec chan<- error

			out, err := NewImageOutput(in, size, outputDir, force)
			if os.IsExist(err) {
			} else if err != nil {
				ec = errc
			} else {
				send = outc
			}

			select {
			case send <- out:
			case ec <- err:
			case <-done:
				return
			}
		}
	}
}

func ResizeImages(done <-chan struct{}, ready <-chan *ImageOutput, resized chan<- *ImageOutput) {
	for in := range ready {
		in.Transform()
		select {
		case resized <- in:
		case <-done:
			return
		}
	}

}

func SaveImages(done <-chan struct{}, errc chan<- error, ready <-chan *ImageOutput, imageQuality int) {
	for in := range ready {
		var ec chan<- error
		var err error

		err = in.Save(imageQuality)
		if err != nil {
			ec = errc
		}
		select {
		case ec <- err:
		case <-done:
			return
		default: // to continue
		}
	}
}
