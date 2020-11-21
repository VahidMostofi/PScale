package autoscaler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/aggregator"
)

type simpleAutoscaler struct {
	MonitorInterval time.Duration // in seconds
	ac              aggregator.RequestCounts
}

// Autoscale ...
func (sa *simpleAutoscaler) Autoscale(signals chan Signal, close chan bool, errs chan error) {
	fmt.Println("autoscaling with monitor interval of", sa.MonitorInterval, "seconds")
	go sa.monitorRequestCounts(close, errs)
}

// monitor ...
func (sa *simpleAutoscaler) monitorRequestCounts(close chan bool, errs chan error) {
	ticker := time.NewTicker(sa.MonitorInterval)
	for {
		select {
		case <-ticker.C:
			end := time.Now().UnixNano() / 1e6
			start := end - sa.MonitorInterval.Nanoseconds()/1e6
			rc, err := sa.ac.GetRequestsCounts(start, end)
			if err != nil {
				errs <- err
				return
			}
			total := 0
			rNames, err := sa.ac.GetRequestsNames(start, end)
			if err != nil {
				errs <- err
				return
			}
			str := "("
			// convertor := int(sa.MonitorInterval.Seconds())
			for i, reqName := range rNames {
				total += int(rc[reqName] / int(sa.MonitorInterval.Seconds()))
				str += strconv.Itoa(rc[reqName] / int(sa.MonitorInterval.Seconds()))

				if i != len(rNames)-1 {
					str += ","
				}
			}
			str += ")"
			str += strconv.Itoa(total)
			fmt.Println(str)
		case <-close:
			fmt.Println("signal to stop simple autoscaling")
			return
		}
	}
}

// GetNewSimpleAutoscaler ...
func GetNewSimpleAutoscaler() (Autoscaler, error) {
	s := &simpleAutoscaler{}
	s.MonitorInterval = viper.GetDuration("autoscale_interval") * time.Second
	_, ac, _, err := aggregator.GetAll()
	if err != nil {
		panic(err)
	}
	s.ac = ac
	return s, nil
}
