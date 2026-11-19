/*
Package v1 provides the client and common types for the Selectel Compute SDK.
The package version is independent of the Compute API version.

The client uses a regional Compute (Nova) endpoint and a project-scoped token supplied by the
caller. Every request carries the Compute API microversion 2.93 of the Selectel fork. The client
does not authenticate, retry requests, or wait for resource state changes. API operations are
provided by resource packages such as github.com/selectel/compute-go/pkg/v1/server.
*/
package v1
