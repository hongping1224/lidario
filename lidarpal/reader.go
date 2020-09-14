package lidarpal

import (
	"math"

	"github.com/hongping1224/lidario"
)

//Reader read Pointcloud
type Reader struct {
	las *lidario.LasFile
}

//NewReader Create a new Reader
func NewReader(las *lidario.LasFile) *Reader {
	return &Reader{las: las}
}

//Read point into channel
func (read *Reader) Read(input chan<- lidario.LasPointer, from, to int) {
	for i := from; i < to; i++ {
		//fmt.Printf("\r%d%%", int((float32(i-from)/float32(to-from))*100))
		p, err := read.las.LasPoint(i)
		if err != nil {
			break
		}
		input <- p
	}
	close(input)
}

//Serve read concerrently
func (read *Reader) Serve(input chan<- lidario.LasPointer, from, to int) {
	go read.Read(input, from, to)
}

//CalculatePartSize return the size of each part for concurent running
func (read *Reader) CalculatePartSize(PartCount int) int {
	partSize := int(math.Ceil(float64(read.las.Header.NumberPoints) / float64(PartCount)))
	return partSize
}
