// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// The destination image file, always a jpg, resized to fit the provided size,
// named from the original filename with the provided suffix.
type ImageOutput struct {
	original, modified image.Image

	outFile   *os.File
	imageSize ImageSize
}

// Opens a new file with a name generated from the size and in filename values.
// If the file already exists in the output directory and force is false, returns
// an error; if force is true, returns the existing file which will be overwritten
// when saved.
func NewImageOutput(in *ImageInput, size ImageSize, outputDir string, force bool) (*ImageOutput, error) {
	filename := filepath.Join(filepath.Clean(outputDir), func(name, suffix string) string {
		ext := filepath.Ext(name)
		return strings.TrimSuffix(name, ext) + suffix + ".jpg"
	}(filepath.Base(in.Filename), size.Suffix))

	flag := os.O_RDWR | os.O_CREATE | os.O_EXCL // no overwriting
	if force {
		flag = os.O_RDWR | os.O_CREATE // open file (truncate once we know we have something to save)
	}
	outFile, err := os.OpenFile(filename, flag, 0666)
	if err != nil {
		return nil, err
	}

	out := new(ImageOutput)
	out.Init(in.Clone().Image, outFile, size)

	return out, nil
}

func (o *ImageOutput) Init(i image.Image, out *os.File, size ImageSize) {
	o.original = i
	o.outFile = out
	o.imageSize = size
}

// Does the image transformation work. In this case, a resize to fit operation.
// This is an independent method to allow effective pipelining.
func (o *ImageOutput) Transform() {
	if o.imageSize.Size != 0 { // do not resize if the size is zero
		o.modified = imaging.Fit(o.original, int(o.imageSize.Size), int(o.imageSize.Size), imaging.Lanczos)
	} else {
		o.modified = o.original
	}
}

// Saves the modified image to the intended output file at the specified quality
func (o *ImageOutput) Save(quality int) error {
	opts := new(jpeg.Options)
	opts.Quality = quality
	if err := o.outFile.Truncate(0); err != nil {
		return fmt.Errorf("Error Saving ImageOutput %v: %v", *o, err)
	}
	if err := jpeg.Encode(o.outFile, o.modified, opts); err != nil {
		return fmt.Errorf("Error Saving ImageOutput %v: %v", *o, err)
	}
	return nil
}

func (o *ImageOutput) Close() error {
	if err := o.outFile.Close(); err != nil {
		return fmt.Errorf("Error Closing ImageOutput: %v", err)
	}
	return nil
}
