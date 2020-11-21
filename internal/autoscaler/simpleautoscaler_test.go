package autoscaler

import "testing"

func TestSimpleAutoscaler(t *testing.T) {
	a, err := GetNewSimpleAutoscaler()
	if err != nil {
		panic(err)
	}
	sigCh := make(chan Signal)
	errCh := make(chan error)
	closeCh := make(chan bool)
	a.Autoscale(sigCh, closeCh, errCh)
	err = <-errCh
	panic(err)
}
