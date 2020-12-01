package autoscaler

// Signal ...
type Signal struct{}

// Autoscaler ...
type Autoscaler interface {
	Autoscale(chan Signal, chan bool, chan error)
}
