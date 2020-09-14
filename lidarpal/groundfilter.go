package lidarpal

import (
	"fmt"

	"gonum.org/v1/gonum/stat"
)

//FindPoleCenter filter out ground, noise to find pole and calculate center point
func FindPoleCenter(input string, voxelCellSize float64, minpoint int, queryDist float64) (gx, gy, gz float64, err error) {
	las, err := OpenLasFile(input)
	if err != nil {
		return
	}
	defer las.Close()
	grid := GenerateVoxel(las, voxelCellSize)

	grid.Filter(minpoint, queryDist)

	centerx, centery, centerz := grid.FindCenterPoint()
	if len(centerx) == 0 {
		return 0, 0, 0, fmt.Errorf("Center not Found on %s", input)
	}
	meanx := stat.Mean(centerx, nil)
	meany := stat.Mean(centery, nil)
	minz := centerz[0]
	for _, v := range centerz {
		if v < minz {
			minz = v
		}
	}
	gx, gy, gz = grid.IndexToGlobal(meanx, meany, minz)
	//fmt.Println(gx, gy, gz)
	return
}

//RemoveGround remove ground by using Voxelize method
func RemoveGround(input, outputPath string, voxelCellSize float64, columethreshold int) (err error) {
	las, err := OpenLasFile(input)
	if err != nil {
		return
	}
	defer las.Close()
	grid := GenerateVoxel(las, voxelCellSize)
	grid.RemoveBottomZVoxel()
	grid.RemoveColumeByCount(columethreshold)
	err = grid.SaveLas(outputPath, las)
	if err != nil {
		return
	}
	return
}
