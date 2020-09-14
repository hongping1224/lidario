package lidarpal

import (
	"fmt"

	"github.com/hongping1224/lidario"
)

//Splitter copy one input point into several output channel
type Splitter struct {
	Input  chan lidario.LasPointer
	Output []chan lidario.LasPointer
}

//NewSplitter create splitter with output channel
func NewSplitter(input chan lidario.LasPointer, numOfOutput int) *Splitter {
	output := make([]chan lidario.LasPointer, numOfOutput)
	for i := 0; i < numOfOutput; i++ {
		output[i] = make(chan lidario.LasPointer)
	}
	return &Splitter{Input: input, Output: output}
}

//Serve start Splitter in goroutine
// stop when input channel is close
func (spl *Splitter) Serve() {
	go func() {
		for {
			c, open := <-spl.Input
			if open == false {
				fmt.Println("Spliter Closing")
				for _, oc := range spl.Output {
					close(oc)
				}
				break
			}
			for _, oc := range spl.Output {
				oc <- c
			}
		}
	}()
}
