package autoscaler

import (
	"fmt"
	"time"

	"github.com/vahidmostofi/wise-auto-scaler/internal/aggregator"
)

// Autoscaler ...
type Autoscaler struct {
}

// Autoscale ...
func Autoscale() {
	_, ac, _, err := aggregator.GetAll()
	if err != nil {
		panic(err)
	}
	var d time.Duration = 10
	forever := make(chan bool)
	ticker := time.NewTicker(d * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				end := time.Now().UnixNano() / 1e6
				start := end - d.Nanoseconds()/1e6
				counts, _ := ac.GetRequestsCounts(start, end)
				total := 0
				for _, value := range counts {
					total += value
				}
				fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
				fmt.Println(total, counts)
				fmt.Println("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
				// case <-closeCh:
				// 	return
			}
		}
	}()
	<-forever
}
