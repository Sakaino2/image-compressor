package controllers

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"path/filepath"
	"strings"

	"golang.org/x/image/bmp"
)

func DecodeImage(file io.Reader, inputPath string) (*image.Image, error) {
	var img image.Image
	var err error
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(file)
	case ".png":
		img, err = png.Decode(file)
	case ".bmp":
		img, err = bmp.Decode(file)
	default:
		img, _, err = image.Decode(file)
	}

	if err != nil {
		return nil, err
	}

	return &img, nil
}
