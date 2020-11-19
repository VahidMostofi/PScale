package evaluator

import (
	"fmt"
	"strconv"
	"time"

	"github.com/montanaflynn/stats"
)

// SLA ...
type SLA interface {
	Check(Evaluator, map[string]string) (bool, error) // returns true if the SLA holds
}

// ResponseTimeSLA ...
type ResponseTimeSLA struct {
	RequestName string
	Property    string // mean, 90p, 95p, 99p
	Threshold   float64
}

// Check ...
func (rSLA *ResponseTimeSLA) Check(e Evaluator, info map[string]string) (bool, error) {
	end := time.Now().UnixNano() / 1e6
	start := end - e.GetIntervalSeconds().Nanoseconds()/1e6
	info["start"] = strconv.FormatInt(start/1e3, 10)
	info["end"] = strconv.FormatInt(end/1e3, 10)
	responseTimes, err := e.GetAggResponseTimes().GetRequestResponseTimes(rSLA.RequestName, start, end)
	if err != nil {
		return false, err
	}
	info["count"] = strconv.Itoa(len(responseTimes))
	if len(responseTimes) == 0 {
		return true, nil
	}
	var value float64 = 0

	if rSLA.Property == "p95" {
		value, err = stats.Percentile(responseTimes, 95)
		if err != nil {
			panic(err)
		}
	} else {
		panic(fmt.Errorf("I don't know %s", rSLA.Property))
	}

	info["value"] = strconv.FormatFloat(value, 'f', 3, 64)

	if value > rSLA.Threshold {
		return false, nil
	}

	return true, nil
}

// ResponseCodeSLA ...
type ResponseCodeSLA struct {
	MinNotOK    int
	RequestName string
}

// Check ...
func (cSLA *ResponseCodeSLA) Check(e Evaluator, info map[string]string) (bool, error) {
	end := time.Now().UnixNano() / 1e6
	start := end - e.GetIntervalSeconds().Nanoseconds()/1e6
	info["start"] = strconv.FormatInt(start/1e3, 10)
	info["end"] = strconv.FormatInt(end/1e3, 10)
	c, err := e.GetAggStatus().GetFailCount(cSLA.RequestName, start, end)
	if err != nil {
		return false, err
	}

	// info["minNotOk"] = strconv.Itoa(cSLA.MinNotOK)
	info["notOkCount"] = strconv.Itoa(c)

	if c > cSLA.MinNotOK {
		return false, nil
	}

	return true, nil
}
