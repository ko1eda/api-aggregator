package http

import (
	koala "github.com/ko1eda/apiaggregator"
)

// This service allows us to abstract away all our lower level json conversions
// All of these items will eventually make external api calls, so they will be blocking.
type AsyncProvider interface {
	GetProviderInfo() (*koala.ProviderInfo, error)
	GetFullMenu() (*koala.Menu, error)
}
