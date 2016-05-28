// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
)

// Represents an image read from a file. Because image.Image may be a pointer value,
// these are not safe to pass by value.
type ImageInput struct {
	Image    image.Image
	Filename string
}

func NewImageInput(filename string) (*ImageInput, error) {
	img, err := imaging.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("ImageInput Error: Error opening %q, %v", filename, err)
	}
	ii := new(ImageInput)
	ii.Filename = filename
	ii.Image = img
	return ii, nil
}

func (ii *ImageInput) Clone() *ImageInput {
	n := new(ImageInput)
	n.Filename = ii.Filename
	n.Image = imaging.Clone(ii.Image)
	return n
}
