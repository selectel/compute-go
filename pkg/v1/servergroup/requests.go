package servergroup

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const (
	basePath = "os-server-groups"
	envelope = "server_group"
)

func Create(ctx context.Context, client *v1.Client, opts CreateOpts) (*View, *v1.Response, error) {
	if opts.Name == "" || opts.Policy == "" {
		return nil, nil, invalid("the server group name and policy are required")
	}

	body := createBody{Name: opts.Name, Policy: opts.Policy}
	if opts.MaxServerPerHost > 0 {
		body.Rules = &Rules{MaxServerPerHost: opts.MaxServerPerHost}
	}

	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, basePath, http.StatusOK, createRequest{ServerGroup: body}, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.ServerGroup, response, nil
}

func Get(ctx context.Context, client *v1.Client, groupID string) (*View, *v1.Response, error) {
	path, err := groupPath(groupID)
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

	return &result.ServerGroup, response, nil
}

// List returns the server groups of the project. The API returns them in one response.
func List(ctx context.Context, client *v1.Client) ([]View, *v1.Response, error) {
	result := &listEnvelope{}

	response, err := transport.DoRequest(ctx, client, http.MethodGet, basePath, http.StatusOK, nil, result)
	if err != nil {
		return nil, response, err
	}

	return result.ServerGroups, response, nil
}

func Delete(ctx context.Context, client *v1.Client, groupID string) (*v1.Response, error) {
	path, err := groupPath(groupID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusNoContent, nil, nil)
}

func groupPath(groupID string) (string, error) {
	if groupID == "" {
		return "", invalid("the server group ID is required")
	}

	return basePath + "/" + url.PathEscape(groupID), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
