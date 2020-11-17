package wrapper

import (
	"fmt"
	"testing"
)

func TestListening(t *testing.T) {
	a, err := GetNewWrapperAggregator()
	if err != nil {
		panic(err)
	}
	fmt.Println(a)
}
