package filter

import "math"

type Bilateral struct {
	sigmaR float64
	sigmaX float64
	sigmaY float64
	sigmaZ float64
	t      float64 //Dynamic range of the input image
}

func NewBilateral(sigmaR, sigmaX, sigmaY, sigmaZ float64) *Bilateral {
	return &Bilateral{
		sigmaR: sigmaR,
		sigmaX: sigmaX,
		sigmaY: sigmaY,
		sigmaZ: sigmaZ,
	}
}

func (f *Bilateral) order() []int {
	// min is the minimum R, B or G value from the image, default is math.Inf(-1).
	// max is the maximum R, B or G value from the image, default is math.Inf(1).
	f.t = 0 // t = max - min, so default is 0.
	n := 0
	gamma := 0.5 * math.Pi
	rho := gamma * f.sigmaR

	if f.sigmaR > 1.0/gamma/gamma {
		n = 3
	} else {
		n = int(math.Max(3, math.Ceil(1.0/rho/rho)))
	}

	nchannels := n + 1
	cmin := 0
	cmax := nchannels
	trunc := 0

	return []int{0}
}
