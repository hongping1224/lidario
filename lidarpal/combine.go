package lidarpal

import (
	"sync"

	"github.com/hongping1224/lidario"
)

//Combine a las to output lasFile
func Combine(inputLas *lidario.LasFile, outputLas *Writer, ReaderCount int) {
	reader := NewReader(inputLas)
	partSize := reader.CalculatePartSize(ReaderCount)
	process := func(p lidario.LasPointer) {
		outputLas.Write(p)
	}
	var Wg sync.WaitGroup
	Wg.Add(ReaderCount)
	for i := 0; i < ReaderCount; i++ {
		c := make(chan lidario.LasPointer, 4)
		go reader.Read(c, partSize*i, partSize*(i+1))
		worker := NewPointWorker(c, process, &Wg)
		go worker.Serve()
	}
	Wg.Wait()
}
