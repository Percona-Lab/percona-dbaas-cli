package dp

import (
	"fmt"
	"sync"
	"time"
)

type DotPrinter struct {
	stopChan chan bool
	wg       sync.WaitGroup
}

func New() DotPrinter {
	return DotPrinter{}
}

func (d *DotPrinter) StartPrintDot(message string) {
	d.stopChan = make(chan bool)
	fmt.Print(message)
	d.PrintDot()
}

func (d *DotPrinter) StopPrintDot(message string) {
	d.stopChan <- true
	d.wg.Wait()
	fmt.Print("[" + message + "]")
	fmt.Println()
}

func (d *DotPrinter) PrintDot() {
	stopPrint := false
	go func(stopChan chan bool) {
		d.wg.Add(1)
		for stop := range stopChan {
			if stop {
				stopPrint = true
			}
		}
	}(d.stopChan)
	go func() {
		for {
			if stopPrint {
				d.wg.Done()
				return
			}
			time.Sleep(1 * time.Second)
			fmt.Print(".")
		}
	}()
}
