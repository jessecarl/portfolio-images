// © Copyright 2016 Jesse Allen. All rights reserved
// Released under the MIT license found in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

// Options are the set of configuration flags for the command
type Options struct {
	ImageSizes   []ImageSize
	InputGlob    string
	OutputDir    string
	ForceSave    bool
	ImageQuality int
	WorkerCount  int
}

func NewOptions() *Options {
	return &Options{
		ImageSizes:   make([]ImageSize, 0, 1),
		ImageQuality: 80,
		WorkerCount:  5,
	}
}

func (opt *Options) AddFlags(fs *pflag.FlagSet) {
	fs.VarP(&imageSizeSlice{&opt.ImageSizes}, "image-sizes", "s",
		"[required] Image Sizes to be output. Should be of the form \"suffix:size\" "+
			"where \"suffix\" is added to the base filename when saving this version, and "+
			"\"size\" defines the maximum square dimension (0 for no resize).")
	fs.StringVarP(&opt.InputGlob, "input-file-glob", "i", opt.InputGlob,
		"[required] Glob used to find all images to be resized.")
	fs.StringVarP(&opt.OutputDir, "output-directory", "o", opt.OutputDir,
		"[required] Output directory to save all generated images to.")
	fs.BoolVarP(&opt.ForceSave, "force-save", "f", opt.ForceSave,
		"Force overwrite of existing images in the output directory. Default is false.")
	fs.IntVarP(&opt.ImageQuality, "image-quality", "q", opt.ImageQuality,
		"Image quality 1-100 inclusive.")
	fs.IntVarP(&opt.WorkerCount, "worker-count", "w", 5,
		"Width of the pipeline or number of workers per pipeline stage.")
}

func (opt *Options) Valid() bool {
	return len(opt.ImageSizes) > 0 &&
		opt.InputGlob != "" &&
		opt.OutputDir != "" &&
		opt.ImageQuality > 1 && opt.ImageQuality < 100 &&
		opt.WorkerCount > 1
}

type imageSizeSlice struct {
	imageSizes *[]ImageSize
}

func (ivs *imageSizeSlice) String() string {
	s := ""
	for _, si := range *ivs.imageSizes {
		s += si.String()
	}
	return s
}

func (ivs *imageSizeSlice) Set(value string) error {
	for _, s := range strings.Split(value, ",") {
		iv := ImageSize{}
		if err := iv.Set(s); err != nil {
			return fmt.Errorf("ImageSizeSlice Error: unable to set ImageSize %q, %v", s, err)
		}
		*ivs.imageSizes = append(*ivs.imageSizes, iv)
	}
	return nil
}

func (ivs *imageSizeSlice) Type() string {
	return "suffix:size"
}
