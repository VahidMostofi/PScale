package evaluator

import (
	"testing"
)

func TestEvaluator(t *testing.T) {
	e, err := GetSimpleEvaluator()
	if err != nil {
		panic(err)
	}

	e.AddSLA("login_response_time", &ResponseTimeSLA{
		Property:    "p95",
		RequestName: "login",
		Threshold:   250,
	})

	e.AddSLA("login_status", &ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "login",
	})

	e.AddSLA("get_book_response_time", &ResponseTimeSLA{
		Property:    "p95",
		RequestName: "get_book",
		Threshold:   250,
	})

	e.AddSLA("get_book_status", &ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "get_book",
	})

	e.AddSLA("edit_book_response_time", &ResponseTimeSLA{
		Property:    "p95",
		RequestName: "edit_book",
		Threshold:   250,
	})

	e.AddSLA("edit_book_status", &ResponseCodeSLA{
		MinNotOK:    1,
		RequestName: "edit_book",
	})

	err = e.Evaluate()
	if err != nil {
		panic(err)
	}
}
