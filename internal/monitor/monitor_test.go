package monitor

import (
	"testing"
	"time"
)

func TestEvaluator(t *testing.T) {
	RecordPodDetails("/home/vahid/Desktop/test2.csv", make(chan bool))
	ticker := time.NewTicker(30 * time.Second)
	<-ticker.C
}
