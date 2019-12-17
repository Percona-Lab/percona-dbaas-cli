package dp

import (
	"fmt"
	"sync"
	"time"
)

type DotPrinter struct {
	stop chan string
	wg   sync.WaitGroup
}

func New() DotPrinter {
	return DotPrinter{
		stop: make(chan string),
	}
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
