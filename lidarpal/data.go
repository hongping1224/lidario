package lidarpal

import (
	"github.com/hongping1224/lidario"
)

//Point struct for 3d point
type Point struct {
	*lidario.PointRecord3
	format uint8
	M      float64
}

//NewPoint create point with xyz
func NewPoint(x, y, z float64) *Point {
	return &Point{PointRecord3: &lidario.PointRecord3{PointRecord0: &lidario.PointRecord0{X: x, Y: y, Z: z}}}
}

//NewPointWithData create point with xyz and m
func NewPointWithData(x, y, z, m float64) *Point {
	return &Point{PointRecord3: &lidario.PointRecord3{PointRecord0: &lidario.PointRecord0{X: x, Y: y, Z: z}}, M: m}
}

// Format returns the point format number.
func (p *Point) Format() uint8 {
	return p.format
}

//Box struct for 3d bounding box
type Box struct {
	Max, Min Point
}
