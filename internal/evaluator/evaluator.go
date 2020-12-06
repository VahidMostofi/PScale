package evaluator

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/aggregator"
)

// Evaluator ...
type Evaluator interface {
	Evaluate(errCh chan error, closeCh chan struct{})
	AddSLA(string, SLA) error
	GetAggCounts() aggregator.RequestCounts
	GetAggResponseTimes() aggregator.RequestResponseTimes
	GetAggStatus() aggregator.RequestStatus
	GetIntervalSeconds() time.Duration
	GetViolationCounts() map[string]int
	GetViolationInfo() map[string][]map[string]string
	GetTestCounts() int
	GetRequestCounts() ([]int64, []int)
}

// Simple ...
type Simple struct {
	aggCounts        aggregator.RequestCounts
	aggResponseTimes aggregator.RequestResponseTimes
	aggStatus        aggregator.RequestStatus
	slas             map[string]SLA
	violationCounts  map[string]int
	violationInfo    map[string][]map[string]string
	IntervalSeconds  time.Duration
	testCount        int
	times            []int64
	counts           []int
}

// GetSimpleEvaluator ...
func GetSimpleEvaluator() (Evaluator, error) {
	se := &Simple{}

	var interval time.Duration = viper.GetDuration("evaluate_interval") * time.Second

	ar, ac, as, err := aggregator.GetAll()
	if err != nil {
		return nil, err
	}

	se.aggCounts = ac
	se.aggResponseTimes = ar
	se.aggStatus = as

	se.slas = make(map[string]SLA)

	se.violationCounts = make(map[string]int)
	se.violationInfo = make(map[string][]map[string]string)

	se.IntervalSeconds = interval

	return se, nil
}

// Evaluate ...
func (se *Simple) Evaluate(errCh chan error, closeCh chan struct{}) {
	fmt.Println("start monitoring for evaluation ")
	fmt.Printf("evaluate every %f seconds\n", se.IntervalSeconds.Seconds())

	ticker := time.NewTicker(se.IntervalSeconds)
	go func() {
		fmt.Println("evaluating ...")
		for {
			select {
			case <-ticker.C:
				se.testCount++
				fmt.Printf("evaluating %d SLAs\n", len(se.slas))
				for name, s := range se.slas {
					info := make(map[string]string)
					isOk, err := s.Check(se, info)
					if err != nil {
						errCh <- fmt.Errorf("error while checking SLA %s: %w", name, err)
					}
					if isOk { // meets the SLA
						// fmt.Println("✅", name, info)
					} else {
						se.violationCounts[name]++
						se.violationInfo[name] = append(se.violationInfo[name], info)
						fmt.Println("❌", name, info)
					}
				}
				fmt.Println("=============================")
			case <-closeCh:
				return
			}
		}
	}()

	ticker2 := time.NewTicker(1 * time.Second)
	go func() {
		fmt.Println("evaluator counting requests ...")
		for {
			select {
			case <-ticker2.C:
				end := time.Now().UnixNano() / 1e6
				start := end - 30e3

				total := 0
				counts, _ := se.aggCounts.GetRequestsCounts(start, end)
				for _, c := range counts {
					total += c
				}
				se.times = append(se.times, start)
				se.counts = append(se.counts, total/30)
				// fmt.Println(total / 30)
			case <-closeCh:
				return
			}
		}
	}()
}

// AddSLA ...
func (se *Simple) AddSLA(key string, s SLA) error {
	if _, ok := se.slas[key]; ok {
		return fmt.Errorf("key %s already exists for slas", key)
	}
	se.slas[key] = s
	se.violationCounts[key] = 0
	se.violationInfo[key] = make([]map[string]string, 0)
	return nil
}

// GetAggCounts ...
func (se *Simple) GetAggCounts() aggregator.RequestCounts { return se.aggCounts }

// GetAggResponseTimes ...
func (se *Simple) GetAggResponseTimes() aggregator.RequestResponseTimes { return se.aggResponseTimes }

// GetAggStatus ...
func (se *Simple) GetAggStatus() aggregator.RequestStatus { return se.aggStatus }

// GetIntervalSeconds ...
func (se *Simple) GetIntervalSeconds() time.Duration { return se.IntervalSeconds }

// GetViolationCounts ...
func (se *Simple) GetViolationCounts() map[string]int { return se.violationCounts }

// GetViolationInfo ...
func (se *Simple) GetViolationInfo() map[string][]map[string]string { return se.violationInfo }

// GetTestCounts ...
func (se *Simple) GetTestCounts() int { return se.testCount }

// GetRequestCounts ...
func (se *Simple) GetRequestCounts() ([]int64, []int) { return se.times, se.counts }
