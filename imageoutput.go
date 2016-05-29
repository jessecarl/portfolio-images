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

	outFile   string
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

	// If the file exists but we don't have the force flag, do not proceed
	info, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("Error preparing output file, %q: %v", filename, err)
	} else if err == nil && !force {
		return nil, fmt.Errorf("Skipping file, %q: use force (-f) to force overwrite", filename)
	} else if err == nil && !info.Mode().IsRegular() {
		return nil, fmt.Errorf("Error preparing output file, %q: cannot overwrite with new image", filename)
	}

	out := new(ImageOutput)
	out.Init(in.Clone().Image, filename, size)

	return out, nil
}

func (o *ImageOutput) Init(i image.Image, out string, size ImageSize) {
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

	flag := os.O_RDWR | os.O_CREATE | os.O_TRUNC
	outFile, err := os.OpenFile(o.outFile, flag, 0666)
	defer outFile.Close()
	if err != nil {
		return fmt.Errorf("Error Saving ImageOutput %v: %v", *o, err)
	}

	if err := jpeg.Encode(outFile, o.modified, opts); err != nil {
		return fmt.Errorf("Error Saving ImageOutput %v: %v", *o, err)
	}
	return nil
}
