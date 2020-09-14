package lidarpal

import (
	"sync"

	"github.com/hongping1224/lidario"
)

//PointWorker process Point from input
type PointWorker struct {
	Input   chan lidario.LasPointer
	Process process
	Wg      *sync.WaitGroup
}

type process func(lidario.LasPointer)

//NewPointWorker create a new Processor
func NewPointWorker(input chan lidario.LasPointer, process process, Wg *sync.WaitGroup) *PointWorker {
	return &PointWorker{Input: input, Process: process, Wg: Wg}
}

//Run start worker
func (p *PointWorker) Run() {
	for {
		a, open := <-p.Input
		if open == false {
			//fmt.Println("Processor Closing")
			p.Wg.Done()
			break
		}
		p.Process(a)
		//fmt.Printf("Greeting from %d, %d\n", i, a)
	}
}

//Serve worker in go routine
func (p *PointWorker) Serve() {
	go p.Run()
}
