package flavor

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/pagination"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const (
	basePath   = "flavors"
	detailPath = "flavors/detail"
	envelope   = "flavor"
)

// Create makes a flavor. The API has no operation that changes a flavor afterwards.
func Create(ctx context.Context, client *v1.Client, opts CreateOpts) (*View, *v1.Response, error) {
	if opts.Name == "" {
		return nil, nil, invalid("the flavor name is required")
	}

	if opts.VCPUs < 1 || opts.RAM < 1 || opts.Disk < 0 {
		return nil, nil, invalid("a flavor needs at least one vCPU, 1 MiB of RAM and a non-negative disk size")
	}

	request := createRequest{Flavor: createBody(opts)}
	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, basePath, http.StatusOK, request, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.Flavor, response, nil
}

func Get(ctx context.Context, client *v1.Client, flavorID string) (*View, *v1.Response, error) {
	path, err := flavorPath(flavorID)
	if err != nil {
		return nil, nil, err
	}

	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodGet, path, http.StatusOK, nil, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.Flavor, response, nil
}

// ListDetail returns all pages of the detailed flavor list and never returns a partial result.
func ListDetail(ctx context.Context, client *v1.Client, opts ListOpts) ([]View, error) {
	query := url.Values{}

	if opts.IsPublic != "" {
		query.Set("is_public", opts.IsPublic)
	}

	return pagination.ReadAll[View](ctx, client, detailPath, "flavors", query, opts.Limit, func() *pageEnvelope {
		return &pageEnvelope{}
	})
}

func Delete(ctx context.Context, client *v1.Client, flavorID string) (*v1.Response, error) {
	path, err := flavorPath(flavorID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusAccepted, nil, nil)
}

func flavorPath(flavorID string) (string, error) {
	if flavorID == "" {
		return "", invalid("the flavor ID is required")
	}

	return basePath + "/" + url.PathEscape(flavorID), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
