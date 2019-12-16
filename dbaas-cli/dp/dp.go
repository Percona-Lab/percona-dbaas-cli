package dp

import (
	"fmt"
	"time"
)

type DotPrinter struct {
	stop chan string
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
}

func (d *DotPrinter) start(message string) {
	tckr := time.NewTicker(time.Second * 1)
	defer tckr.Stop()
	fmt.Printf(message)
	for {
		select {
		case <-tckr.C:
			fmt.Print(".")
		case msg := <-d.stop:
			fmt.Printf("[%s]\n", msg)
			return
		}
	}
}
