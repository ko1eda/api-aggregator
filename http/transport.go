package http

import (
	"net/http"
	"time"
)

// We pass this to our providers so we can easily mock out HTTP responses during testing
type HttpGetter interface {
	Get(url string) (*http.Response, error)
}

// Create a new Http Client with a 30 second connection timeout
func NewClient() *http.Client {
	t := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	return &http.Client{Transport: t}
}
