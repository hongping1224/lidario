package lidarpal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hongping1224/lidario"
)

//SplitbySourceID point cloud by sourceID
func SplitbySourceID(laspath, outputPath string, ReaderCount int) {
	var Wg sync.WaitGroup
	las, err := OpenLasFile(laspath)
	if err != nil {
		return
	}
	//Setup processor
	reader := NewReader(las)
	var Writers sync.Map

	process := func(p lidario.LasPointer) {
		//Create Writer as needed
		id := p.PointData().PointSourceID
		val, ok := Writers.Load(id)
		if ok == false {
			basename := strings.TrimSuffix(filepath.Base(laspath), filepath.Ext(laspath))
			err := os.MkdirAll(outputPath, os.ModePerm)
			if err != nil {
				fmt.Println("Fail to create Output Dir")
				panic(err)
			}
			output := filepath.Join(outputPath, fmt.Sprintf("%s_%d.las", basename, id))
			outLas, err := lidario.InitializeUsingFile(output, las)
			if err != nil {
				fmt.Println("Fail to create Las File")
				panic(err)
			}
			w := NewWriter(make(chan lidario.LasPointer))
			w.Serve(outLas)
			Writers.Store(id, w)
			val = w
		}
		val.(*Writer).Write(p)
	}
	//Create Readers
	partSize := reader.CalculatePartSize(ReaderCount)
	Wg.Add(ReaderCount)
	for i := 0; i < ReaderCount; i++ {
		c := make(chan lidario.LasPointer, 2)
		reader.Serve(c, partSize*i, partSize*(i+1))
		worker := NewPointWorker(c, process, &Wg)
		go worker.Serve()
	}
	Wg.Wait()
	//Close all writers channel
	Writers.Range(func(ki, vi interface{}) bool {
		vi.(*Writer).Close()
		return true
	})
	las.Close()
}
