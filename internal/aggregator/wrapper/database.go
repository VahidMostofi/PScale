package wrapper

import (
	"fmt"
)

var minInterval int64 = 100

// Database ..
type Database struct {
	RTDs map[string]*RequestTypeDatabase
}

// GetNewDatabase ...
func GetNewDatabase(requestTypes []string, bucketInterval int64) *Database {
	d := &Database{}
	d.RTDs = make(map[string]*RequestTypeDatabase)
	for _, r := range requestTypes {
		d.RTDs[r] = GetNewRequestTypeDatabase(r, bucketInterval)
	}
	return d
}

// Add ..
func (d *Database) Add(requestType string, r *RequestInfo) int64 {
	return d.RTDs[requestType].Add(r)
}

// Get ...
func (d *Database) Get(requestType string, start, end int64) []*RequestInfo {
	return d.RTDs[requestType].Get(start, end)
}

// RequestTypeDatabase ...
type RequestTypeDatabase struct {
	RequestType    string
	bucketInterval int64
	buckets        map[int64][]*RequestInfo
}

// RequestInfo ...
type RequestInfo struct {
	ResponseTime float64
	Success      bool
	Timestamp    int64
}

// GetNewRequestTypeDatabase ...
func GetNewRequestTypeDatabase(requestType string, bucketInterval int64) *RequestTypeDatabase {
	if bucketInterval < minInterval {
		panic(fmt.Errorf("bucket interval can't be less than minInterval"))
	}
	d := &RequestTypeDatabase{}
	d.RequestType = requestType
	d.bucketInterval = bucketInterval
	d.buckets = make(map[int64][]*RequestInfo)
	return d
}

// Add ...
func (rtd *RequestTypeDatabase) Add(r *RequestInfo) int64 {
	key := r.Timestamp / rtd.bucketInterval
	if _, ok := rtd.buckets[key]; !ok {
		rtd.buckets[key] = make([]*RequestInfo, 0)
	}
	rtd.buckets[key] = append(rtd.buckets[key], r)
	return key
}

// Get ...
func (rtd *RequestTypeDatabase) Get(start, end int64) []*RequestInfo {
	keys := make(map[int64]bool) // to make sure there is no duplicate
	for i := start / rtd.bucketInterval; i <= 1+(end/rtd.bucketInterval); i++ {
		keys[i] = true
	}
	requests := make([]*RequestInfo, 0)

	for key := range keys {
		requests = append(requests, rtd.buckets[key]...)
	}

	return requests
}

func (rtd *RequestTypeDatabase) cleanup() {
	//TODO
}
