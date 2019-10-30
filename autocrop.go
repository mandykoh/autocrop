package autocrop

import (
	"fmt"
	"image"
	"image/draw"
	"math"
)

const bufferPixels = 1

// BoundsForThreshold returns the bounds for a crop of the specified image up to
// the given energy threshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func BoundsForThreshold(img *image.NRGBA, energyThreshold float32) image.Rectangle {

	crop := img.Bounds()
	if crop.Empty() {
		return crop
	}

	// Calculate cumulative energies by column
	energies := make([]float32, crop.Dx(), crop.Dx())
	nextEnergies := make([]float32, crop.Dx(), crop.Dx())
	for i := crop.Min.Y; i < crop.Max.Y; i++ {
		for j := crop.Min.X; j < crop.Max.X; j++ {
			offset := j - crop.Min.X
			nextEnergies[offset] = energy(img, j, i)

			if i > crop.Min.Y {
				nextEnergies[offset] += energies[offset]
			}
		}

		energies, nextEnergies = nextEnergies, energies
	}

	// Find left and right high energy jumps
	maxEnergy := findMaxEnergy(energies)
	cropLeft := findFirstEnergyBound(energies, maxEnergy, energyThreshold)
	cropRight := findLastEnergyBound(energies, maxEnergy, energyThreshold, cropLeft)

	// Calculate cumulative energies by row
	energies = make([]float32, crop.Dy(), crop.Dy())
	nextEnergies = make([]float32, crop.Dy(), crop.Dy())
	for j := crop.Min.X; j < crop.Max.X; j++ {
		for i := crop.Min.Y; i < crop.Max.Y; i++ {
			offset := i - crop.Min.Y
			nextEnergies[offset] = energy(img, j, i)

			if j > crop.Min.X {
				nextEnergies[offset] += energies[offset]
			}
		}

		energies, nextEnergies = nextEnergies, energies
	}

	// Find top and bottom high energy jumps
	maxEnergy = findMaxEnergy(energies)
	cropTop := findFirstEnergyBound(energies, maxEnergy, energyThreshold)
	cropBottom := findLastEnergyBound(energies, maxEnergy, energyThreshold, cropTop)

	// Apply the crop
	crop.Min.X += cropLeft
	crop.Min.Y += cropTop
	crop.Max.X -= cropRight
	crop.Max.Y -= cropBottom

	return crop
}

// ToThreshold returns an image cropped using the bounds provided by
// BoundsForThreshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func ToThreshold(img *image.NRGBA, energyThreshold float32) *image.NRGBA {
	crop := BoundsForThreshold(img, energyThreshold)
	resultImg := image.NewNRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
	draw.Draw(resultImg, resultImg.Bounds(), img, crop.Min, draw.Src)
	return resultImg
}

func colourAt(img *image.NRGBA, x, y int) (r, g, b, a uint8) {
	c := img.NRGBAAt(x, y)
	return c.R, c.G, c.B, c.A
}

func energy(img *image.NRGBA, x, y int) float32 {
	neighbours := [8]float32{
		luminance(colourAt(img, x-1, y-1)),
		luminance(colourAt(img, x, y-1)),
		luminance(colourAt(img, x+1, y-1)),
		luminance(colourAt(img, x-1, y)),
		luminance(colourAt(img, x+1, y)),
		luminance(colourAt(img, x-1, y+1)),
		luminance(colourAt(img, x, y+1)),
		luminance(colourAt(img, x+1, y+1)),
	}

	eX := neighbours[0] + neighbours[3] + neighbours[5] - neighbours[2] - neighbours[4] - neighbours[7]
	eY := neighbours[0] + neighbours[1] + neighbours[2] - neighbours[5] - neighbours[6] - neighbours[7]

	return float32(math.Abs(float64(eX)) + math.Abs(float64(eY))*(float64(img.NRGBAAt(x, y).A)/255))
}

func findFirstEnergyBound(energies []float32, maxEnergy, threshold float32) (bound int) {
	for i := 1; i < len(energies); i++ {
		if energies[i]/maxEnergy >= threshold {
			bound = i + bufferPixels
			break
		}
		bound++
	}

	return bound
}

func findLastEnergyBound(energies []float32, maxEnergy, threshold float32, firstBound int) (bound int) {
	for i := len(energies) - 2; i > firstBound+bufferPixels; i-- {
		if energies[i]/maxEnergy >= threshold {
			bound = len(energies) - i - 1 + bufferPixels
			break
		}
		bound++
	}

	return bound
}

func findMaxEnergy(energies []float32) float32 {
	max := energies[0]
	for i := 1; i < len(energies); i++ {
		if energies[i] > max {
			max = energies[i]
		}
	}

	return max
}

func luminance(r, g, b, a uint8) float32 {
	return 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b) + float32(a)
}
