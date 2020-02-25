package pb

import (
	"fmt"
	"sync"
	"time"
)

type ProgressBar interface {
	Start(message string)
	Stop(message string)
}

type DotPrinter struct {
	stop chan string
	wg   sync.WaitGroup
}

func NewDotPrinter() *DotPrinter {
	dp := DotPrinter{
		stop: make(chan string),
	}

	return &dp
}

func (d *DotPrinter) Start(message string) {
	go d.start(message)
}

func (d *DotPrinter) Stop(message string) {
	d.stop <- message
	d.wg.Wait()
}

func (d *DotPrinter) start(message string) {
	d.wg.Add(1)
	tckr := time.NewTicker(time.Second * 5)
	defer tckr.Stop()
	fmt.Printf(message)
	for {
		select {
		case <-tckr.C:
			fmt.Print(".")
		case msg := <-d.stop:
			fmt.Printf("[%s]\n", msg)
			d.wg.Done()
			return
		}
	}
}

type NoOp struct{}

func NewNoOp() *NoOp {
	no := NoOp{}

	return &no
}

func (n *NoOp) Start(message string) {}

func (n *NoOp) Stop(message string) {}
