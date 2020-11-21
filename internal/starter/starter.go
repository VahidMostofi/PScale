package starter

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/vahidmostofi/wise-auto-scaler/internal/autoscaler"
	"github.com/vahidmostofi/wise-auto-scaler/internal/evaluator"
	"gopkg.in/yaml.v2"
)

// EvaluationReport ...
type EvaluationReport struct {
	Start           int64                          `yaml:"start"`
	End             int64                          `yaml:"end"`
	ViolationCounts map[string]int                 `yaml:"violationsCounts"`
	ViolationInfo   map[string][]map[string]string `yaml:"violationsInfo"`
}

// StartEvaluator ...
func StartEvaluator(done chan bool) {
	var ReportPath = viper.GetString("evaluate_report_path")
	const SystemName = "bookstore-report"
	er := EvaluationReport{}

	startEvaluationTime := time.Now().Unix()
	finishEvaluationTime := time.Now().Unix()

	er.Start = startEvaluationTime
	er.End = finishEvaluationTime

	e, err := evaluator.GetSimpleEvaluator()
	if err != nil {
		panic(err)
	}

	e.AddSLA("login_response_time", &evaluator.ResponseTimeSLA{
		Property:    "p95",
		RequestName: "login",
		Threshold:   250,
	})

	e.AddSLA("login_status", &evaluator.ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "login",
	})

	e.AddSLA("get_book_response_time", &evaluator.ResponseTimeSLA{
		Property:    "p95",
		RequestName: "get_book",
		Threshold:   250,
	})

	e.AddSLA("get_book_status", &evaluator.ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "get_book",
	})

	e.AddSLA("edit_book_response_time", &evaluator.ResponseTimeSLA{
		Property:    "p95",
		RequestName: "edit_book",
		Threshold:   250,
	})

	e.AddSLA("edit_book_status", &evaluator.ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "edit_book",
	})

	errCh := make(chan error)
	closeCh := make(chan struct{})
	go e.Evaluate(errCh, closeCh)

	if err != nil {
		panic(err)
	}

	// handle interupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println("got", sig.String(), "saving report at:")
			filePath := ReportPath + "as-" + SystemName + "-" + strconv.FormatInt(time.Now().Unix(), 10) + ".yml"
			fmt.Println(filePath)
			er.ViolationInfo = e.GetViolationInfo()
			er.ViolationCounts = e.GetViolationCounts()
			b, err := yaml.Marshal(er)
			if err != nil {
				panic(err)
			}
			err = ioutil.WriteFile(filePath, b, 0777)
			if err != nil {
				panic(err)
			}
			done <- true
			fmt.Println("Evaluator is done")
			return
		}
	}()
}

// StartAutoscaler ...
func StartAutoscaler() {
	a, err := autoscaler.GetNewSimpleAutoscaler()
	if err != nil {
		panic(err)
	}
	sigCh := make(chan autoscaler.Signal)
	errCh := make(chan error)
	closeCh := make(chan bool)
	a.Autoscale(sigCh, closeCh, errCh)

	doEvaluate := viper.GetBool("evaluate_enable")
	evaluateDone := make(chan bool)
	if doEvaluate {
		StartEvaluator(evaluateDone)
	}

	for {
		select {
		case err := <-errCh:
			panic(err)
		case <-evaluateDone:
			return
		}
	}
}
