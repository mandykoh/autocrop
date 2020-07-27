package autocrop

import (
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path"
	"testing"
)

func BenchmarkEnergySummation(b *testing.B) {
	b.StopTimer()
	img := loadTestImage("70x70-pink-square-on-clouds.png", nil)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		BoundsForThreshold(img, 0.01)
	}
}

func TestBoundsForThreshold(t *testing.T) {

	t.Run("returns bounds for simple image with plain background cropped out", func(t *testing.T) {
		img := loadTestImage("70x70-pink-square-on-white.png", t)

		result := BoundsForThreshold(img, 0.01)

		if expected, actual := 70, result.Dx(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels wide but was %d", expected, actual)
		}
		if expected, actual := 70, result.Dy(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels tall but was %d", expected, actual)
		}

		for i := result.Min.Y; i < result.Max.Y; i++ {
			for j := result.Min.X; j < result.Max.X; j++ {
				c := img.NRGBAAt(j, i)

				if expected, actual := (color.NRGBA{R: 228, G: 0, B: 140, A: 255}), c; expected != actual {
					t.Fatalf("Expected entire cropped region to be pink but found %v", actual)
				}
			}
		}
	})

	t.Run("returns bounds for simple image with gradient background cropped out", func(t *testing.T) {
		img := loadTestImage("70x70-pink-square-on-gradient.png", t)

		result := BoundsForThreshold(img, 0.31)

		if expected, actual := 70, result.Dx(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels wide but was %d", expected, actual)
		}
		if expected, actual := 70, result.Dy(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels tall but was %d", expected, actual)
		}

		for i := result.Min.Y; i < result.Max.Y; i++ {
			for j := result.Min.X; j < result.Max.X; j++ {
				c := img.NRGBAAt(j, i)

				if expected, actual := (color.NRGBA{R: 228, G: 0, B: 140, A: 255}), c; expected != actual {
					t.Fatalf("Expected entire cropped region to be pink but found %v", actual)
				}
			}
		}
	})

	t.Run("returns bounds for simple image with textured background cropped out", func(t *testing.T) {
		img := loadTestImage("70x70-pink-square-on-clouds.png", t)

		result := BoundsForThreshold(img, 0.2)

		if expected, actual := 70, result.Dx(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels wide but was %d", expected, actual)
		}
		if expected, actual := 70, result.Dy(); expected != actual {
			t.Errorf("Expected cropped bounds to be %d pixels tall but was %d", expected, actual)
		}

		for i := result.Min.Y; i < result.Max.Y; i++ {
			for j := result.Min.X; j < result.Max.X; j++ {
				c := img.NRGBAAt(j, i)

				if expected, actual := (color.NRGBA{R: 228, G: 0, B: 140, A: 255}), c; expected != actual {
					t.Fatalf("Expected entire cropped region to be pink but found %v", actual)
				}
			}
		}
	})

	t.Run("returns increasingly cropped images for higher thresholds", func(t *testing.T) {
		img := loadTestImage("radial-gradient.png", t)

		lastBounds := img.Bounds()

		thresholds := []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.9}

		for i := range thresholds {
			result := BoundsForThreshold(img, thresholds[i])

			if !result.In(lastBounds) || result.Eq(lastBounds) {
				t.Errorf("Expected cropped bounds for threshold %f to be smaller than %v but was %v", thresholds[i], lastBounds, result)
			}

			lastBounds = result
		}
	})

	t.Run("works with a 1x1 image", func(t *testing.T) {
		img := loadTestImage("1x1.png", t)

		result := BoundsForThreshold(img, 0.01)

		if expected, actual := 1, result.Dx(); expected != actual {
			t.Errorf("Expected cropped bounds to be 1 pixel wide but was %d", actual)
		}
		if expected, actual := 1, result.Dy(); expected != actual {
			t.Errorf("Expected cropped bounds to be 1 pixel tall but was %d", actual)
		}

		for i := result.Min.Y; i < result.Max.Y; i++ {
			for j := result.Min.X; j < result.Max.X; j++ {
				c := img.NRGBAAt(j, i)

				if expected, actual := (color.NRGBA{R: 255, G: 255, B: 255, A: 255}), c; expected != actual {
					t.Fatalf("Expected entire cropped region to be white but found %v", actual)
				}
			}
		}
	})
}

func TestToThreshold(t *testing.T) {

	t.Run("returns complex image with transparent background cropped out", func(t *testing.T) {
		img := loadTestImage("avocado.png", t)

		result := ToThreshold(img, 0.01)

		expectedResult := image.NewNRGBA(image.Rect(0, 0, 424, 549))
		draw.Draw(expectedResult, expectedResult.Bounds(), img, image.Pt(img.Bounds().Min.X+48, img.Bounds().Min.Y+71), draw.Src)

		if expected, actual := expectedResult.Bounds().Dx(), result.Bounds().Dx(); expected != actual {
			t.Errorf("Expected cropped result to be %d pixels wide but was %d", expected, actual)
		}
		if expected, actual := expectedResult.Bounds().Dy(), result.Bounds().Dy(); expected != actual {
			t.Errorf("Expected cropped result to be %d pixels tall but was %d", expected, actual)
		}

		for i := result.Bounds().Min.Y; i < result.Bounds().Max.Y; i++ {
			for j := result.Bounds().Min.X; j < result.Bounds().Max.X; j++ {
				c := result.NRGBAAt(j, i)

				if expected, actual := expectedResult.NRGBAAt(j-result.Bounds().Min.X, i-result.Bounds().Min.Y), c; expected != actual {
					t.Fatalf("Found difference in cropped result at (%d,%d)", j, i)
				}
			}
		}
	})
}

func loadTestImage(fileName string, t *testing.T) *image.NRGBA {
	if t != nil {
		t.Helper()
	}

	inFile, err := os.Open(path.Join("test-images", fileName))
	if err != nil {
		if t != nil {
			t.Fatalf("%v", err)
		} else {
			return nil
		}
	}
	defer inFile.Close()

	img, _, err := image.Decode(inFile)
	if err != nil {
		if t != nil {
			t.Fatalf("%v", err)
		} else {
			return nil
		}
	}

	nrgbaImg := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
	draw.Draw(nrgbaImg, nrgbaImg.Bounds(), img, img.Bounds().Min, draw.Src)

	return nrgbaImg
}
