package autocrop

import (
	"image"
	"image/draw"
	"math"
)

// BoundsForThreshold returns the bounds for a crop of the specified image up to
// the given energy threshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func BoundsForThreshold(img *image.NRGBA, energyThreshold float32) image.Rectangle {
	const bufferPixels = 1

	crop := img.Bounds()

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

	// Find max column energy
	maxEnergy := float32(0.0)
	for i := 0; i < len(energies); i++ {
		if energies[i] > maxEnergy {
			maxEnergy = energies[i]
		}
	}

	// Find high energy jump from left
	cropLeft := 0
	for i := 1; i < len(energies); i++ {
		if energies[i]/maxEnergy >= energyThreshold {
			cropLeft = i + bufferPixels
			break
		}
		cropLeft++
	}

	// Find high energy jump from right
	cropRight := 0
	for i := len(energies) - 2; i > cropLeft+bufferPixels; i-- {
		if energies[i]/maxEnergy >= energyThreshold {
			cropRight = len(energies) - i - 1 + bufferPixels
			break
		}
		cropRight++
	}

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

	// Find max row energy
	maxEnergy = 0.0
	for i := 0; i < len(energies); i++ {
		if energies[i] > maxEnergy {
			maxEnergy = energies[i]
		}
	}

	// Find high energy jump from top
	cropTop := 0
	for i := 1; i < len(energies); i++ {
		if energies[i]/maxEnergy >= energyThreshold {
			cropTop = i + bufferPixels
			break
		}
		cropTop++
	}

	// Find high energy jump from bottom
	cropBottom := 0
	for i := len(energies) - 2; i > cropTop+bufferPixels; i-- {
		if energies[i]/maxEnergy >= energyThreshold {
			cropBottom = len(energies) - i - 1 + bufferPixels
			break
		}
		cropBottom++
	}

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

func colourAt(img *image.NRGBA, x, y int) (r, g, b uint8) {
	c := img.NRGBAAt(x, y)
	return c.R, c.G, c.B
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

func luminance(r, g, b uint8) float32 {
	return 0.2126*float32(r) + 0.7152*float32(g) + 0.0722*float32(b)
}
