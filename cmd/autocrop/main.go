package main

import (
	"fmt"
	"github.com/mandykoh/autocrop"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"os"
	"strconv"
)

func main() {

	if len(os.Args) < 3 {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: autocrop <image_path> <output_png_path> [energy_threshold]\n")
		os.Exit(1)
	}

	imgFilePath := os.Args[1]
	imgFile, err := os.Open(imgFilePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not open image file: %v\n", err)
		os.Exit(2)
	}
	defer imgFile.Close()

	outFilePath := os.Args[2]

	img, _, err := image.Decode(imgFile)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not read image: %v\n", err)
		os.Exit(2)
	}

	threshold := float32(0.01)
	if len(os.Args) > 3 {
		val, err := strconv.ParseFloat(os.Args[3], 32)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "invalid threshold: %v\n", err)
			os.Exit(1)
		}

		threshold = float32(val)
	}

	nrgbaImg := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(nrgbaImg, nrgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)

	result := autocrop.ToThreshold(nrgbaImg, threshold)

	outFile, err := os.Create(outFilePath)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not create output file: %v\n", err)
		os.Exit(2)
	}
	defer outFile.Close()

	err = png.Encode(outFile, result)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "could not write output image: %v\n", err)
		os.Exit(2)
	}
}
