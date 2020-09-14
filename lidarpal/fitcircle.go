package lidarpal

import "math"

func bar(in []float64) float64 {
	sum := 0.0
	for _, v := range in {
		sum = sum + v
	}
	return sum / float64(len(in))
}

func calcSuu(in []float64) float64 {
	sum := 0.0
	for _, v := range in {
		sum = sum + (v * v)
	}
	return sum
}

func calcSuuu(in []float64) float64 {
	sum := 0.0
	for _, v := range in {
		sum = sum + (v * v * v)
	}
	return sum
}

func calcSuv(in1 []float64, in2 []float64) float64 {
	sum := 0.0
	for k := range in1 {
		sum = sum + in1[k]*in2[k]
	}
	return sum
}

func calcSuvv(in1 []float64, in2 []float64) float64 {
	sum := 0.0
	for k := range in1 {
		sum = sum + in1[k]*in2[k]*in2[k]
	}
	return sum
}

func calcU(in []float64, bar float64) []float64 {
	out := make([]float64, len(in))

	for k, v := range in {
		out[k] = v - bar
	}
	return out
}

// CalcLeastSquareCircleFit computes a least square fit circle for a list of 2d-coordinates.
// It takes the x and y coordinates as arguments. Obviously the two
// argument arrays must have the same length.
// The function returns three values: The x,y location of the circle center
// and the radius of the circle.
func CalcLeastSquareCircleFit(x []float64, y []float64) (float64, float64, float64) {

	N := len(x)

	xbar := bar(x)
	ybar := bar(y)

	u := calcU(x, xbar)
	v := calcU(y, ybar)

	suu := calcSuu(u)
	suv := calcSuv(u, v)
	svv := calcSuu(v)

	suuu := calcSuuu(u)
	svvv := calcSuuu(v)

	suvv := calcSuvv(u, v)
	svuu := calcSuvv(v, u)

	e4 := 0.5 * (suuu + suvv)
	e5 := 0.5 * (svvv + svuu)

	uc := (svv*e4 - suv*e5) / (suu*svv - suv*suv)
	vc := (e4 - uc*suu) / suv

	xc := uc + xbar
	yc := vc + ybar
	r := math.Sqrt(uc*uc + vc*vc + (suu+svv)/float64(N))

	return xc, yc, r
}
