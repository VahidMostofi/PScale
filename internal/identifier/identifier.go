package identifier

import (
	"fmt"
	"strings"
)

// INFOHTTPMethod ...
const INFOHTTPMethod = "HTTPMethod"

// HTTPMethodGET ...
const HTTPMethodGET = "GET"

// HTTPMethodPOST ...
const HTTPMethodPOST = "POST"

// HTTPMethodPUT ...
const HTTPMethodPUT = "PUT"

// INFOHTTPPath ...
const INFOHTTPPath = "HTTPPath"

// Identifier ...
type Identifier interface {
	GetType(map[string]string) (string, error)
	GetTypes() []string
}

type regexIdentifier struct {
}

func (r *regexIdentifier) GetType(info map[string]string) (string, error) {
	if info[INFOHTTPMethod] == HTTPMethodPOST && strings.HasPrefix(info[INFOHTTPPath], "/auth/login") {
		return "login", nil
	} else if info[INFOHTTPMethod] == HTTPMethodGET && strings.HasPrefix(info[INFOHTTPPath], "/books") {
		return "get_book", nil
	} else if info[INFOHTTPMethod] == HTTPMethodPUT && strings.HasPrefix(info[INFOHTTPPath], "/books") {
		return "edit_book", nil
	} else {
		return "", fmt.Errorf("unknown request type based on %v", info)
	}

}

func (r *regexIdentifier) GetTypes() []string {
	return []string{"login", "get_book", "edit_book"}
}

// GetNewIdentifier ...
func GetNewIdentifier() (Identifier, error) {
	return &regexIdentifier{}, nil
}
