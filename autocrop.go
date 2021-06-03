package autocrop

import (
	"github.com/mandykoh/prism/srgb"
	"image"
	"image/draw"
	"math"
)

const bufferPixels = 2

// BoundsForThreshold returns the bounds for a crop of the specified image up to
// the given energy threshold.
//
// energyThreshold is a value between 0.0 and 1.0 representing the maximum
// energy to allow to be cropped away before stopping, relative to the maximum
// energy of the image.
func BoundsForThreshold(img *image.NRGBA, energyThreshold float32) image.Rectangle {

	crop := img.Bounds()

	energyCrop := crop
	energyCrop.Min.X++
	energyCrop.Min.Y++
	energyCrop.Max.X--
	energyCrop.Max.Y--

	if energyCrop.Empty() {
		return img.Bounds()
	}

	colEnergies, rowEnergies := Energies(img, energyCrop)

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

// Energies returns the total row and column energies for the specified region
// of an image.
func Energies(img *image.NRGBA, r image.Rectangle) (cols, rows []float32) {
	imageWidth := img.Rect.Dx()

	luminances, alphas := luminancesAndAlphas(img)

	cols = make([]float32, r.Dx(), r.Dx())
	rows = make([]float32, r.Dy(), r.Dy())

	// Calculate total column and row energies
	for i, row := r.Min.Y, 0; i < r.Max.Y; i, row = i+1, row+1 {
		for j, col := r.Min.X, 0; j < r.Max.X; j, col = j+1, col+1 {
			e := energy(imageWidth, luminances, alphas, j, i)
			cols[col] += e
			rows[row] += e
		}
	}

	return cols, rows
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

func colourAt(img *image.NRGBA, x, y int) (col srgb.Color, alpha float32) {
	return srgb.ColorFromNRGBA(img.NRGBAAt(x, y))
}

func energy(width int, luminances, alphas []float32, x, y int) float32 {
	center := y*width + x

	// North west + west + south west - north east - east - south east
	eX := luminances[center-width-1] + luminances[center-1] + luminances[center+width-1] - luminances[center-width+1] - luminances[center+1] - luminances[center+width+1]

	// North west + north + north east - south west - south - south east
	eY := luminances[center-width-1] + luminances[center-width] + luminances[center-width+1] - luminances[center+width-1] - luminances[center+width] - luminances[center+width+1]

	return float32((math.Abs(float64(eX)) + math.Abs(float64(eY))) * float64(alphas[center]))
}

func findFirstEnergyBound(energies []float32, maxEnergy, threshold float32) (bound int) {
	for i := 0; i < len(energies); i++ {
		if energies[i]/maxEnergy > threshold {
			bound = i
			if i > 0 {
				bound += bufferPixels
			}
			break
		}
		bound++
	}

	return bound
}

func findLastEnergyBound(energies []float32, maxEnergy, threshold float32, firstBound int) (bound int) {
	for i := len(energies) - 1; i > firstBound+bufferPixels; i-- {
		if energies[i]/maxEnergy > threshold {
			bound = len(energies) - i - 1
			if i < len(energies)-1 {
				bound += bufferPixels
			}
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

func luminance(c srgb.Color, alpha float32) float32 {
	return c.Luminance() + alpha
}

func luminancesAndAlphas(img *image.NRGBA) (luminances, alphas []float32) {
	luminances = make([]float32, img.Rect.Dx()*img.Rect.Dy(), img.Rect.Dx()*img.Rect.Dy())
	alphas = make([]float32, img.Rect.Dx()*img.Rect.Dy(), img.Rect.Dx()*img.Rect.Dy())

	offset := 0

	// Get the luminances and alphas for all pixels
	for i, row := img.Rect.Min.Y, 0; i < img.Rect.Max.Y; i, row = i+1, row+1 {
		for j, col := img.Rect.Min.X, 0; j < img.Rect.Max.X; j, col = j+1, col+1 {
			c, a := colourAt(img, col, row)
			luminances[offset] = luminance(c, a)
			alphas[offset] = a
			offset++
		}
	}

	return luminances, alphas
}
