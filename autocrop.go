package autocrop

import (
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

	colEnergies := make([]float32, crop.Dx(), crop.Dx())
	rowEnergies := make([]float32, crop.Dy(), crop.Dy())

	// Calculate total column and row energies
	for i, row := crop.Min.Y, 0; i < crop.Max.Y; i, row = i+1, row+1 {
		for j, col := crop.Min.X, 0; j < crop.Max.X; j, col = j+1, col+1 {
			e := energy(img, j, i)
			colEnergies[col] += e
			rowEnergies[row] += e
		}
	}

	// Find left and right high energy jumps
	maxEnergy := findMaxEnergy(colEnergies)
	cropLeft := findFirstEnergyBound(colEnergies, maxEnergy, energyThreshold)
	cropRight := findLastEnergyBound(colEnergies, maxEnergy, energyThreshold, cropLeft)

	// Find top and bottom high energy jumps
	maxEnergy = findMaxEnergy(rowEnergies)
	cropTop := findFirstEnergyBound(rowEnergies, maxEnergy, energyThreshold)
	cropBottom := findLastEnergyBound(rowEnergies, maxEnergy, energyThreshold, cropTop)

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
