package transport

import "net/url"

type requestOptions struct {
	query            url.Values
	responseEnvelope string
	alsoExpected     []int
}

type RequestOption func(*requestOptions)

func newRequestOptions(options []RequestOption) *requestOptions {
	result := &requestOptions{}

	for _, option := range options {
		option(result)
	}

	return result
}

func WithQuery(query url.Values) RequestOption {
	return func(options *requestOptions) {
		options.query = query
	}
}

func WithResponseEnvelope(name string) RequestOption {
	return func(options *requestOptions) {
		options.responseEnvelope = name
	}
}

// WithAlsoExpectedStatus accepts one more success status for operations that answer differently
// depending on the prior state, such as adding a tag that the server already has.
func WithAlsoExpectedStatus(status int) RequestOption {
	return func(options *requestOptions) {
		options.alsoExpected = append(options.alsoExpected, status)
	}
}

func (options *requestOptions) acceptsStatus(status int) bool {
	for _, expected := range options.alsoExpected {
		if expected == status {
			return true
		}
	}

	return false
}
