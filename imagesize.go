// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// ImageSize is used to couple a destination image size with a destination image name
type ImageSize struct {
	Suffix string
	Size   uint
}

// ImageSizes are represeted as "suffix:size"
func (iv ImageSize) String() string {
	return fmt.Sprintf("%s:%d", iv.Suffix, iv.Size)
}

// ImageSizes are parsed from strings of the form "suffix:size"
func (iv *ImageSize) Set(value string) error {
	raw := strings.Split(value, ":")
	if len(raw) != 2 {
		return fmt.Errorf("ImageSize Error: %q is in incorrect format, must be suffix:size", value)
	}

	suffix := raw[0]
	for _, r := range suffix { // no control characters allowed in filenames
		if unicode.IsControl(r) {
			return fmt.Errorf("ImageSize Error: suffix %q contains control characters", suffix)
		}
	}

	size, err := strconv.ParseUint(raw[1], 10, 64)
	if err != nil {
		return fmt.Errorf("ImageSize Error: parsing size %q, %v", raw[1], err)
	}
	iv.Suffix = suffix
	iv.Size = uint(size)
	return nil
}

// So we can parse multiple image versions from flags
type imageSizeSlice []ImageSize

func (ivs *imageSizeSlice) String() string {
	s := ""
	for _, si := range *ivs {
		s += si.String()
	}
	return s
}

func (ivs *imageSizeSlice) Set(value string) error {
	for _, s := range strings.Split(value, ",") {
		iv := ImageSize{}
		if err := iv.Set(s); err != nil {
			return fmt.Errorf("ImageSizeSlice Error: unable to set ImageSize %q, %v", s, err)
		} else {
			*ivs = append(*ivs, iv)
		}
	}
	return nil
}
