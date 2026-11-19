# compute-go: Go SDK for the Selectel Compute API

Package `compute-go` provides a Go SDK to work with the Selectel public Compute API,
which is the Selectel fork of the OpenStack Nova API.

The SDK package version `v1` is the version of the SDK public interface. It does
not identify the remote API version; every request carries the Compute API microversion 2.93.

The SDK accepts a regional Compute endpoint, a project-scoped token, a `User-Agent`
prefix and an optional HTTP client. It does not authenticate, retry requests or wait for
resource state changes.

## Installation

```sh
go get github.com/selectel/compute-go
```

## Usage

```go
client, err := v1.NewClient(v1.Config{
	Endpoint: "https://ru-9.cloud.api.selcloud.ru/compute/v2.1",
	Token:    token,
})
```
