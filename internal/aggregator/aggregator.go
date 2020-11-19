package aggregator

import "github.com/vahidmostofi/wise-auto-scaler/internal/aggregator/wrapper"

// RequestResponseTimes ...
type RequestResponseTimes interface {
	GetRequestsNames(int64, int64) ([]string, error)
	GetRequestResponseTimes(string, int64, int64) ([]float64, error)
	GetRequestsResponseTimes(int64, int64) (map[string][]float64, error)
	Done() error
	AllDone() error
}

// RequestCounts ...
type RequestCounts interface {
	GetRequestsNames(int64, int64) ([]string, error)
	GetRequestCount(string, int64, int64) (int, error)
	GetRequestsCounts(int64, int64) (map[string]int, error)
	Done() error
	AllDone() error
}

// RequestStatus ...
type RequestStatus interface {
	GetRequestsNames(int64, int64) ([]string, error)
	GetFailCount(string, int64, int64) (int, error)
	GetFailCounts(int64, int64) (map[string]int, error)
	Done() error
	AllDone() error
}

// GetAll returns RequestResponseTimes, RequestCounts, RequestStatus, error
func GetAll() (RequestResponseTimes, RequestCounts, RequestStatus, error) {
	w, err := wrapper.GetNewWrapperAggregator()
	if err != nil {
		return nil, nil, nil, err
	}

	return w, w, w, nil

}
