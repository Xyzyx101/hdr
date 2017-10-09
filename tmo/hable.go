package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/Xyzyx101/hdr"
)

// Hable is an implementation of the Uncharted 2 tonempper by John Hable
// http://filmicworlds.com/blog/filmic-tonemapping-operators/
type Hable struct {
	HDRImage     hdr.Image
	ExposureBias float64
	Gamma        float64
}

// NewDefaultHable returns a hable tone mapper with default settings
func NewDefaultHable(img hdr.Image) *Hable {
	return NewHable(img, 2.0, 2.2)
}

// NewHable instantates a hable tone mapper
func NewHable(img hdr.Image, exposureBias, gamma float64) *Hable {
	return &Hable{
		HDRImage:     img,
		ExposureBias: exposureBias,
		Gamma:        gamma,
	}
}

// Perform the tonemaping operation
func (t *Hable) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())
	t.tonemap(img)
	// for x := 0; x < 100; x++ {
	// 	fmt.Print(img.At(x, 0))
	// }
	return img
}

func (t *Hable) tonemap(img *image.RGBA64) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			inPix := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := inPix.HDRRGBA()
			img.SetRGBA64(x, y, color.RGBA64{
				R: t.hable(r),
				G: t.hable(g),
				B: t.hable(b),
				A: RangeMax,
			})
		}
	}
}

func (t *Hable) hable(x float64) uint16 {
	const (
		A          = 0.15
		B          = 0.50
		C          = 0.10
		D          = 0.20
		E          = 0.02
		F          = 0.30
		W          = 11.2
		WhiteScale = ((W*(A*W+C*B) + D*E) / (W*(A*W+B) + D*F)) - E/F
	)
	x *= 16.0 * t.ExposureBias
	col := ((x*(A*x+C*B) + D*E) / (x*(A*x+B) + D*F)) - E/F
	col *= WhiteScale
	col = math.Pow(col, 1.0/t.Gamma)
	col *= 1 << 16
	return uint16(col)
}
