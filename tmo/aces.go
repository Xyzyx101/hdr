package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/Xyzyx101/hdr"
	"github.com/go-gl/mathgl/mgl64"
)

// ACES is an implementation of the Academy Color Encoding System http://www.oscars.org/science-technology/sci-tech-projects/aces
// This is a simple version I got from here https://github.com/TheRealMJP/BakingLab/blob/master/BakingLab/ACES.hlsl
type ACES struct {
	HDRImage     hdr.Image
	ExposureBias float64
}

// NewDefaultACES return an ACES tmo with default gamma of 1.8
func NewDefaultACES(img hdr.Image) *ACES {
	return NewACES(img, 1.8)
}

// NewACES instantates a ACES tone mapper
func NewACES(img hdr.Image, exposureBias float64) *ACES {
	return &ACES{
		HDRImage:     img,
		ExposureBias: exposureBias,
	}
}

// Perform the tonemaping operation
func (t *ACES) Perform() image.Image {
	img := image.NewRGBA64(t.HDRImage.Bounds())
	t.tonemap(img)
	return img
}

func (t *ACES) tonemap(img *image.RGBA64) {
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			inPix := t.HDRImage.HDRAt(x, y)
			r, g, b, _ := inPix.HDRRGBA()
			colorVec := acesFitted(mgl64.Vec3{r, g, b})
			rgb := linearToRGB(colorVec.Mul(t.ExposureBias))
			img.SetRGBA64(x, y, color.RGBA64{
				R: rgb[0],
				G: rgb[1],
				B: rgb[2],
				A: RangeMax,
			})
		}
	}
}

// Matrices is transposed from the example to work with the matrix library

// var acesInputMat = mgl64.Mat3{
// 	0.59719, 0.35458, 0.04823,
// 	0.07600, 0.90834, 0.01566,
// 	0.02840, 0.13383, 0.83777,
// }

var acesInputMat = mgl64.Mat3{
	0.59719, 0.07600, 0.02840,
	0.35458, 0.90834, 0.13383,
	0.04823, 0.01566, 0.83777,
}

// var acesOutputMat = mgl64.Mat3{
// 	1.60475, -0.53108, -0.07367,
// 	-0.10208, 1.10813, -0.00605,
// 	-0.00327, -0.07276, 1.07602,
// }

var acesOutputMat = mgl64.Mat3{
	1.60475, -0.10208, -0.00327,
	-0.53108, 1.10813, -0.07276,
	-0.07367, -0.00605, 1.07602,
}

func acesFitted(color mgl64.Vec3) mgl64.Vec3 {
	color = acesInputMat.Mul3x1(color)
	color = rttAnODTFit(color)
	color = acesOutputMat.Mul3x1(color)
	color = saturate(color)
	return color
}

func rttAnODTFit(x mgl64.Vec3) mgl64.Vec3 {
	a := x.Add(mgl64.Vec3{0.0245786, 0.0245786, 0.0245786})
	a[0] *= x[0]
	a[1] *= x[1]
	a[2] *= x[2]
	a = a.Add(mgl64.Vec3{-0.000090537, -0.000090537, -0.000090537})

	b := x.Mul(0.983729)
	b = b.Add(mgl64.Vec3{0.4329510, 0.4329510, 0.4329510})
	b[0] *= x[0]
	b[1] *= x[1]
	b[2] *= x[2]
	b = b.Add(mgl64.Vec3{0.238081, 0.238081, 0.238081})

	return mgl64.Vec3{a[0] / b[0], a[1] / b[1], a[2] / b[2]}
}

func saturate(in mgl64.Vec3) mgl64.Vec3 {
	return mgl64.Vec3{
		math.Max(0.0, math.Min(in[0], 1.0)),
		math.Max(0.0, math.Min(in[1], 1.0)),
		math.Max(0.0, math.Min(in[2], 1.0)),
	}
}

const linCurve = 1.0 / 2.4

func linearToRGB(linear mgl64.Vec3) []uint16 {
	x := linear.Mul(12.92)

	y := saturate(linear)
	y = mgl64.Vec3{
		math.Pow(y[0], linCurve),
		math.Pow(y[1], linCurve),
		math.Pow(y[2], linCurve),
	}
	y = y.Mul(1.055)
	y = y.Add(mgl64.Vec3{-0.055, -0.055, -0.055})
	rgb := make([]uint16, 3, 3)
	for i := range rgb {
		if linear[i] < 0.0031308 {
			rgb[i] = uint16(x[i] * RangeMax)
		} else {
			rgb[i] = uint16(y[i] * RangeMax)
		}
	}
	return rgb
}
