package lidarpal

import (
	"fmt"
	"math"
	"sync"

	"github.com/hongping1224/lidario"
)

//VoxelGrid store point index in Grid
type VoxelGrid struct {
	XSize, YSize, ZSize int
	CellSize            float64
	MinX, MinY, MinZ    float64
	Index               [][][][]int // [x][y][z][len]
}

//CreateVoxelGrid of x y z
func CreateVoxelGrid(XSize, YSize, ZSize int, MinX, MinY, MinZ float64, CellSize float64) *VoxelGrid {
	arr := make([][][][]int, XSize)
	for i := range arr {
		arr[i] = make([][][]int, YSize)
		for j := range arr[i] {
			arr[i][j] = make([][]int, ZSize)
			for k := range arr[i][j] {
				arr[i][j][k] = make([]int, 0)
			}
		}
	}
	grid := VoxelGrid{XSize: XSize, YSize: YSize, ZSize: ZSize,
		MinX: MinX, MinY: MinY, MinZ: MinZ,
		CellSize: CellSize, Index: arr}
	return &grid
}

//GenerateVoxel from Lasfile
func GenerateVoxel(las *lidario.LasFile, size float64) *VoxelGrid {
	XSize := int(math.Ceil((las.Header.MaxX-las.Header.MinX)/size)) + 1
	YSize := int(math.Ceil((las.Header.MaxY-las.Header.MinY)/size)) + 1
	ZSize := int(math.Ceil((las.Header.MaxZ-las.Header.MinZ)/size)) + 1
	grid := CreateVoxelGrid(XSize, YSize, ZSize, las.Header.MinX, las.Header.MinY, las.Header.MinZ, size)

	for i := 0; i < las.Header.NumberPoints; i++ {
		//fmt.Printf("\r%d%%", int((float32(i-from)/float32(to-from))*100))
		x, y, z, err := las.GetXYZ(i)
		if err != nil {
			continue
		}
		xindex := int(math.Floor((x - grid.MinX) / size))
		yindex := int(math.Floor((y - grid.MinY) / size))
		zindex := int(math.Floor((z - grid.MinZ) / size))

		grid.Index[xindex][yindex][zindex] = append(grid.Index[xindex][yindex][zindex], i)
	}
	return grid
}

//SaveLas save index to las file at path
func (grid *VoxelGrid) SaveLas(path string, oriLas *lidario.LasFile) error {
	pointChannel := make(chan lidario.LasPointer, 100)
	writer := NewWriter(pointChannel)
	outLas, err := lidario.InitializeUsingFile(path, oriLas)
	if err != nil {
		fmt.Println(path)
		fmt.Println(err)
		return fmt.Errorf("Fail to output File at %s", path)
	}
	writer.Serve(outLas)
	var wg sync.WaitGroup
	for x := 0; x < grid.XSize; x++ {
		for y := 0; y < grid.YSize; y++ {
			for z := 0; z < grid.ZSize; z++ {
				wg.Add(1)
				go func(x, y, z int, output chan<- lidario.LasPointer) {
					defer wg.Done()
					for i := 0; i < len(grid.Index[x][y][z]); i++ {
						p, err := oriLas.LasPoint(grid.Index[x][y][z][i])
						if err != nil {
							continue
						}
						output <- p
					}
				}(x, y, z, pointChannel)
			}
		}
	}
	wg.Wait()
	writer.Close()
	outLas.Close()
	return nil
}

//IndexToGlobal calculate global coordinate form index
func (grid *VoxelGrid) IndexToGlobal(x, y, z float64) (gx, gy, gz float64) {

	gx = (x * grid.CellSize) + grid.MinX
	gy = (y * grid.CellSize) + grid.MinY
	gz = (z * grid.CellSize) + grid.MinZ
	return
}

//FindCenterPoint return each layer center point
func (grid *VoxelGrid) FindCenterPoint() (xa, ya, za []float64) {
	xa = make([]float64, 0)
	ya = make([]float64, 0)
	za = make([]float64, 0)

	for z := 0; z < grid.ZSize; z++ {
		x := make([]float64, 0)
		y := make([]float64, 0)
		for i := 0; i < grid.XSize; i++ {
			for j := 0; j < grid.YSize; j++ {
				if len(grid.Index[i][j][z]) == 0 {
					continue
				}
				x = append(x, float64(i))
				y = append(y, float64(j))
			}
		}
		if len(x) == 0 {
			continue
		}
		xc, yc, _ := CalcLeastSquareCircleFit(x, y)
		if math.IsNaN(xc) || math.IsNaN(yc) {
			continue
		}
		//fmt.Println(r)
		xa = append(xa, xc)
		ya = append(ya, yc)
		za = append(za, float64(z))
	}
	return
}

//RemoveBottomZVoxel check for lowest voxel in each xy  and remove it
func (grid *VoxelGrid) RemoveBottomZVoxel() {
	var wg sync.WaitGroup
	for x := 0; x < grid.XSize; x++ {
		for y := 0; y < grid.YSize; y++ {
			wg.Add(1)
			go func(x, y int) {
				for z := 0; z < grid.ZSize; z++ {
					if len(grid.Index[x][y][z]) != 0 {
						grid.Index[x][y][z] = make([]int, 0)
						break
					}
				}
				wg.Done()
			}(x, y)
		}
	}
	wg.Wait()
}

//RemoveColumeByCount check for each xy colume and remove if total point is lower than threshold
func (grid *VoxelGrid) RemoveColumeByCount(threshold int) {
	Count := make([][]int, grid.XSize)
	for i := range Count {
		Count[i] = make([]int, grid.YSize)
	}
	var wg sync.WaitGroup
	for x := 0; x < grid.XSize; x++ {
		for y := 0; y < grid.YSize; y++ {
			wg.Add(1)
			go func(x, y int) {
				for z := 0; z < grid.ZSize; z++ {
					Count[x][y] += len(grid.Index[x][y][z])
				}
				wg.Done()
			}(x, y)
		}
	}
	wg.Wait()
	for x := 0; x < grid.XSize; x++ {
		for y := 0; y < grid.YSize; y++ {
			wg.Add(1)
			go func(x, y int) {
				defer wg.Done()
				if Count[x][y] > threshold {
					return
				}
				for z := 0; z < grid.ZSize; z++ {
					grid.Index[x][y][z] = make([]int, 0)
				}

			}(x, y)
		}
	}
	wg.Wait()
}

//Filter out outside Point layer by layer using dbscan to find cluster
// if more than 1 cluster, cluster closest to center in picked if same distance, cluster with more point is pick
func (grid *VoxelGrid) Filter(minPoint int, queryDistance float64) {
	var wg sync.WaitGroup
	for z := 0; z < grid.ZSize; z++ {
		wg.Add(1)
		//fmt.Println("filter Z:", z)
		go func(z int) {
			label, class := grid.dbScan(z, minPoint, queryDistance)
			if class > 1 {
				//find best class
				//fmt.Println("find best Z:", z, class)
				class = bestclass(class, z, label, grid)
				//fmt.Println("done find best Z:", z, class)
			}
			for i := 0; i < grid.XSize; i++ {
				for j := 0; j < grid.YSize; j++ {
					if label[i][j] != class {
						//if not best class remove point index
						grid.Index[i][j][z] = make([]int, 0)
					}
				}
			}
			wg.Done()
		}(z)
	}
	wg.Wait()
}
func bestclass(class, z int, label [][]int, grid *VoxelGrid) int {
	bestclass := 1
	mindistance := math.MaxFloat64
	coor := make([][][2]int, class+1)
	bbox := make([][4]int, class+1)
	for i := range coor {
		coor[i] = make([][2]int, 0)
		bbox[i] = [4]int{0, math.MaxInt32, 0, math.MaxInt32}
	}
	for i := 0; i < grid.XSize; i++ {
		for j := 0; j < grid.YSize; j++ {
			c := label[i][j]
			if c <= 0 {
				continue
			}

			coor[c] = append(coor[c], [2]int{i, j})
			xmax, xmin, ymax, ymin := bbox[c][0], bbox[c][1], bbox[c][2], bbox[c][3]
			xmax = max(i, xmax)
			xmin = min(i, xmin)
			ymax = max(j, ymax)
			ymin = min(j, ymin)
			bbox[c] = [4]int{xmax, xmin, ymax, ymin}
		}
	}
	for c := 1; c < class; c++ {
		xmax, xmin, ymax, ymin := bbox[c][0], bbox[c][1], bbox[c][2], bbox[c][3]
		dist := distBetweenCell(grid.XSize/2, grid.YSize/2, 0, (xmax-xmin)/2, (ymax-ymin)/2, 0)
		if dist < mindistance {
			bestclass = c
			mindistance = dist
		} else if dist == mindistance && bestclass != c {
			pa, pb := 0, 0
			for _, cor := range coor[c] {
				pa += len(grid.Index[cor[0]][cor[1]][z])
			}
			for _, cor := range coor[bestclass] {
				pb += len(grid.Index[cor[0]][cor[1]][z])
			}
			if pa > pb {
				bestclass = c
			}
		}
	}
	return bestclass
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func (grid *VoxelGrid) dbScan(z, minPoint int, queryDistance float64) (label [][]int, class int) {
	class = 0

	label = make([][]int, grid.XSize)
	for i := range label {
		label[i] = make([]int, grid.YSize)
	}

	for i := 0; i < grid.XSize; i++ {
		for j := 0; j < grid.YSize; j++ {
			if len(grid.Index[i][j][z]) == 0 {
				continue
			}
			if label[i][j] != 0 {
				continue
			}
			neighbor, ncount := grid.rangeQuery2D(i, j, z, queryDistance)
			if ncount < minPoint {
				label[i][j] = -1
				continue
			}
			class++
			label[i][j] = class
			seedPoint := neighbor
			for len(seedPoint) != 0 {
				x, y := seedPoint[0][0], seedPoint[0][1]
				seedPoint = seedPoint[1:]
				if label[x][y] == -1 {
					label[x][y] = class
				}
				if label[x][y] != 0 {
					continue
				}
				label[x][y] = class
				neigh, c := grid.rangeQuery2D(x, y, z, queryDistance)
				if c >= minPoint {
					seedPoint = append(seedPoint, neigh...)
				}
			}
		}
	}
	return
}

func (grid *VoxelGrid) rangeQuery2D(x, y, z int, queryDistance float64) (neighbor [][2]int, count int) {
	count = 0
	neighbor = make([][2]int, 0)
	maxCell := queryDistance / grid.CellSize
	minx := x - int(math.Ceil(maxCell))
	maxx := x + int(math.Ceil(maxCell))
	miny := y - int(math.Ceil(maxCell))
	maxy := y + int(math.Ceil(maxCell))
	minx = int(math.Min(0, float64(minx)))
	miny = int(math.Min(0, float64(miny)))
	maxx = int(math.Max(float64(grid.XSize), float64(maxx)))
	maxy = int(math.Max(float64(grid.YSize), float64(maxy)))

	for i := minx; i < maxx; i++ {
		for j := miny; j < maxy; j++ {
			if grid.checkInRange(i, j, z) == false {
				continue
			}
			if len(grid.Index[i][j][z]) == 0 {
				continue
			}
			if distBetweenCell(x, y, z, i, j, z) < maxCell {
				count += len(grid.Index[i][j][z])
				neighbor = append(neighbor, [2]int{i, j})
			}
		}
	}
	return
}

func (grid *VoxelGrid) checkInRange(x, y, z int) bool {
	if x < 0 || y < 0 || z < 0 {
		return false
	}
	if x >= grid.XSize || y >= grid.YSize || z >= grid.ZSize {
		return false
	}
	return true
}

func distBetweenCell(x1, y1, z1, x2, y2, z2 int) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	dz := float64(z1 - z2)
	return math.Sqrt(math.Pow(dx, 2) + math.Pow(dy, 2) + math.Pow(dz, 2))
}
