package starter

import (
	"testing"
)

func TestEvaluator(t *testing.T) {
	evaluatorIsDone := make(chan bool)
	StartEvaluator(evaluatorIsDone)
	<-evaluatorIsDone
}
