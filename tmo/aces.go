package tmo

import (
	"image"
	"image/color"
	"math"

	"github.com/Xyzyx101/hdr"
	"gonum.org/v1/gonum/mat"
)

// ACES is an implementation of the Academy Color Encoding System http://www.oscars.org/science-technology/sci-tech-projects/aces
// This is a simple version I got from here https://github.com/TheRealMJP/BakingLab/blob/master/BakingLab/ACES.hlsl
type ACES struct {
	HDRImage hdr.Image
	Gamma    float64
}

// NewDefaultACES return an ACES tmo with default gamma of 1.8
func NewDefaultACES(img hdr.Image) *ACES {
	return NewACES(img, 1.8)
}

// NewACES instantates a ACES tone mapper
func NewACES(img hdr.Image, gamma float64) *ACES {
	return &ACES{
		HDRImage: img,
		Gamma:    gamma,
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
			colorVec := acesFitted(rgbToVec(r, g, b))
			rgb := t.linearToRGB(colorVec)
			img.SetRGBA64(x, y, color.RGBA64{
				R: rgb[0],
				G: rgb[1],
				B: rgb[2],
				A: RangeMax,
			})
		}
	}
}

var acesInputMat = mat.NewDense(3, 3, []float64{
	0.59719, 0.35458, 0.04823,
	0.07600, 0.90834, 0.01566,
	0.02840, 0.13383, 0.83777,
})

var acesOutputMat = mat.NewDense(3, 3, []float64{
	1.60475, -0.53108, -0.07367,
	-0.10208, 1.10813, -0.00605,
	-0.00327, -0.07276, 1.07602,
})

func rgbToVec(r, g, b float64) *mat.VecDense {
	return mat.NewVecDense(3, []float64{r, g, b})
}

func acesFitted(color *mat.VecDense) *mat.VecDense {
	color.MulVec(acesInputMat, color)
	color = rttAnODTFit(color)
	color.MulVec(acesOutputMat, color)
	color = saturate(color)
	return color
}

func rttAnODTFit(x *mat.VecDense) *mat.VecDense {
	var a mat.VecDense
	a.AddVec(x, mat.NewVecDense(3, []float64{0.0245786, 0.0245786, 0.0245786}))
	a.MulElemVec(x, &a)
	a.AddVec(&a, mat.NewVecDense(3, []float64{-0.000090537, -0.000090537, -0.000090537}))
	//b := mat.NewVecDense(3, []float64{0, 0, 0})
	var b mat.VecDense
	b.ScaleVec(0.983729, x)
	b.AddVec(&b, mat.NewVecDense(3, []float64{0.4329510, 0.4329510, 0.4329510}))
	b.MulElemVec(x, &b)
	b.AddVec(&b, mat.NewVecDense(3, []float64{0.238081, 0.238081, 0.238081}))
	a.DivElemVec(&a, &b)
	//return mat.NewVecDense(3, []float64{a.At(0, 0) / b.At(0, 0), a.At(1, 0) / b.At(1, 0), a.At(2, 0) / b.At(2, 0)})
	return &a
}

func saturate(x *mat.VecDense) *mat.VecDense {
	x.SetVec(0, math.Max(0.0, math.Min(x.At(0, 0), 1.0)))
	x.SetVec(1, math.Max(0.0, math.Min(x.At(1, 0), 1.0)))
	x.SetVec(2, math.Max(0.0, math.Min(x.At(2, 0), 1.0)))
	return x
}

func (t *ACES) linearToRGB(linear *mat.VecDense) []uint16 {
	linear.ScaleVec(t.Gamma, linear)
	var x mat.VecDense
	x.ScaleVec(12.95, linear)
	linear = saturate(linear)
	y := mat.NewVecDense(3, []float64{
		math.Pow(linear.At(0, 0), 1.0/2.4),
		math.Pow(linear.At(1, 0), 1.0/2.4),
		math.Pow(linear.At(2, 0), 1.0/2.4),
	})
	y.ScaleVec(1.055, y)
	y.AddVec(y, mat.NewVecDense(3, []float64{-0.055, -0.055, -0.055}))
	rgb := make([]uint16, 3, 3)
	for i := range rgb {
		if linear.At(i, 0) < 0.0031308 {
			rgb[i] = uint16(x.At(i, 0) * RangeMax)
		} else {
			rgb[i] = uint16(y.At(i, 0) * RangeMax)
		}
	}
	return rgb
}
