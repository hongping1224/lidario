package lidarpal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hongping1224/lidario"
)

//Crop point cloud
func Crop(lasPath []string, radius float64, intensityThreshold uint16, minPoint int, signs []Point, outputRootPath string) {

	for i, las := range lasPath {
		fmt.Printf("Cropping File : %s(%d/%d)\n", las, i+1, len(lasPath))
		lasfile, err := OpenLasFile(las)
		if err != nil {
			continue
		}
		lasfile.SetFixedRadiusSearchDistance(radius, true)
		if err != nil {
			continue
		}
		for _, sign := range signs {
			results := lasfile.FixedRadiusSearch3D(sign.PointRecord0.X, sign.PointRecord0.Y, sign.PointRecord0.Z)
			if results.Len() == 0 {
				continue
			}
			points := make([]lidario.LasPointer, 0)
			for results.Len() > 0 {
				result, err := results.Pop()
				if err != nil {
					fmt.Println(err)
					continue
				}
				p, err := lasfile.LasPoint(result.Index)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if p.PointData().Intensity >= intensityThreshold {

					points = append(points, p)
				}
			}
			if len(points) <= minPoint {
				continue
			}
			directory := filepath.Join(outputRootPath, fmt.Sprintf("%d", int(sign.M)))
			if _, err := os.Stat(directory); os.IsNotExist(err) {
				os.MkdirAll(directory, os.ModePerm)
			}
			savePath := filepath.Join(directory, filepath.Base(las))
			outLas, err := lidario.InitializeUsingFile(savePath, lasfile)
			if err != nil {
				return
			}
			outLas.AddLasPoints(points)
			outLas.Close()
		}
		lasfile.Close()
	}

}

//GetNumOfPoint return the number of point in las
func GetNumOfPoint(fileName string) (int, error) {
	las, err := OpenLasHeader(fileName)
	if err != nil {
		return 0, err
	}
	nop := las.Header.NumberPoints
	las.Close()
	return nop, nil
}

//GetHeader return the header of las file
func GetHeader(fileName string) (lidario.LasHeader, error) {
	las, err := OpenLasHeader(fileName)
	if err != nil {
		return lidario.LasHeader{}, err
	}
	header := las.Header
	las.Close()
	return header, nil
}

//OpenLasHeader open las as read header mode
func OpenLasHeader(fileName string) (*lidario.LasFile, error) {
	return lidario.NewLasFile(fileName, "rh")
}

//OpenLasFile open las as read mode
func OpenLasFile(fileName string) (*lidario.LasFile, error) {
	return lidario.NewLasFile(fileName, "r")
}

//BoxOverlap check box b1 overlaps b2
func BoxOverlap(b1, b2 Box) bool {
	if between(b1.Min.X, b1.Max.X, b2.Min.X, b2.Max.X) == false {
		return false
	}
	if between(b1.Min.Y, b1.Max.Y, b2.Min.Y, b2.Max.Y) == false {
		return false
	}
	if between(b1.Min.Z, b1.Max.Z, b2.Min.Z, b2.Max.Z) == false {
		return false
	}
	return true
}

func between(amin, amax, bmin, bmax float64) bool {
	if amin <= bmin && bmin <= amax {
		return true
	}
	if amin <= bmax && bmax <= amax {
		return true
	}
	if bmin <= amin && amin <= bmax {
		return true
	}
	if bmin <= amax && amax <= bmax {
		return true
	}
	return false
}

//PointInBox check point in box
func PointInBox(pt lidario.LasPointer, b Box) bool {
	p := pt.PointData()
	fmt.Println("Point", p)
	fmt.Println("box", b.Min.PointRecord0, b.Max.PointRecord0)
	if p.X < b.Min.X {
		return false
	}
	if p.X > b.Max.X {
		return false
	}
	if p.Y < b.Min.Y {
		return false
	}
	if p.Y > b.Max.Y {
		return false
	}
	if p.Z < b.Min.Z {
		return false
	}
	if p.Z > b.Max.Z {
		return false
	}
	return true
}

//PointBuffer create a square buffer around Point
func PointBuffer(p, shift Point, r float64) (Point, Point) {
	return *NewPoint(p.X+shift.X-r, p.Y+shift.Y-r, p.Z+shift.Z-r), *NewPoint(p.X+shift.X+r, p.Y+shift.Y+r, p.Z+shift.Z+r)
}

func checkSignInLas(header lidario.LasHeader, signs []Point) []int {
	minPoint := NewPoint(header.MinX, header.MinY, header.MinZ)
	maxPoint := NewPoint(header.MaxX, header.MaxY, header.MaxZ)
	boundingBox := Box{Min: *minPoint, Max: *maxPoint}
	index := make([]int, 0)
	for i, sign := range signs {
		if PointInBox(sign.PointRecord0, boundingBox) {
			index = append(index, i)
		}
	}
	return index
}
