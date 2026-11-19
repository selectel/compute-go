package server

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/pagination"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const (
	basePath   = "servers"
	detailPath = "servers/detail"
	envelope   = "server"
)

// Create asks the cloud to build a server and returns its ID. A structured fault with a 4xx status
// means that nothing was created; any other failure leaves the outcome unknown, because the API
// has no idempotency key.
func Create(ctx context.Context, client *v1.Client, opts CreateOpts) (*Created, *v1.Response, error) {
	request, err := opts.request()
	if err != nil {
		return nil, nil, err
	}

	result := &createdEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, basePath, http.StatusAccepted, request, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.Server, response, nil
}

func Get(ctx context.Context, client *v1.Client, serverID string) (*View, *v1.Response, error) {
	path, err := serverPath(serverID)
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

	return &result.Server, response, nil
}

// ListDetail returns all pages of the detailed server list and never returns a partial result.
func ListDetail(ctx context.Context, client *v1.Client, opts ListOpts) ([]View, error) {
	query := url.Values{}

	if opts.Name != "" {
		query.Set("name", opts.Name)
	}

	if opts.Status != "" {
		query.Set("status", opts.Status)
	}

	return pagination.ReadAll[View](ctx, client, detailPath, "servers", query, opts.Limit, func() *pageEnvelope {
		return &pageEnvelope{}
	})
}

// Update changes the name and the user data in one request, without a restart of the server.
func Update(ctx context.Context, client *v1.Client, serverID string, opts UpdateOpts) (*View, *v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	if opts.isEmpty() {
		return nil, nil, invalid("an update needs the name or the user data")
	}

	if opts.UserData != nil && opts.ClearUserData {
		return nil, nil, invalid("the user data cannot be both set and cleared")
	}

	body := map[string]any{}

	if opts.Name != nil {
		body["name"] = *opts.Name
	}

	if opts.UserData != nil {
		body["user_data"] = *opts.UserData
	}

	if opts.ClearUserData {
		body["user_data"] = nil
	}

	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPut, path, http.StatusOK, updateRequest{Server: body}, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.Server, response, nil
}

func Delete(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusNoContent, nil, nil)
}

// SetMetadata merges the given keys into the server metadata and returns the visible metadata.
// Keys that are not given are kept, including the service keys that the API hides.
func SetMetadata(
	ctx context.Context, client *v1.Client, serverID string, metadata map[string]string,
) (map[string]string, *v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	if len(metadata) == 0 {
		return nil, nil, invalid("at least one metadata key is required")
	}

	result := &metadataEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, path+"/metadata", http.StatusOK, metadataEnvelope{Metadata: metadata}, result,
	)
	if err != nil {
		return nil, response, err
	}

	return result.Metadata, response, nil
}

func DeleteMetadatum(ctx context.Context, client *v1.Client, serverID, key string) (*v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, err
	}

	if key == "" {
		return nil, invalid("the metadata key is required")
	}

	return transport.DoRequest(
		ctx, client, http.MethodDelete, path+"/metadata/"+url.PathEscape(key), http.StatusNoContent, nil, nil,
	)
}

// AddTag adds one tag and keeps the others. Adding a tag that the server already has succeeds.
func AddTag(ctx context.Context, client *v1.Client, serverID, tag string) (*v1.Response, error) {
	path, err := tagPath(serverID, tag)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(
		ctx, client, http.MethodPut, path, http.StatusCreated, nil, nil,
		transport.WithAlsoExpectedStatus(http.StatusNoContent),
	)
}

func DeleteTag(ctx context.Context, client *v1.Client, serverID, tag string) (*v1.Response, error) {
	path, err := tagPath(serverID, tag)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusNoContent, nil, nil)
}

func ListTags(ctx context.Context, client *v1.Client, serverID string) ([]string, *v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	result := &tagsEnvelope{}

	response, err := transport.DoRequest(ctx, client, http.MethodGet, path+"/tags", http.StatusOK, nil, result)
	if err != nil {
		return nil, response, err
	}

	return result.Tags, response, nil
}

// ListInterfaces returns the ports attached to the server in the order the API reports them.
func ListInterfaces(ctx context.Context, client *v1.Client, serverID string) ([]Interface, *v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	result := &interfacesEnvelope{}

	response, err := transport.DoRequest(ctx, client, http.MethodGet, path+"/os-interface", http.StatusOK, nil, result)
	if err != nil {
		return nil, response, err
	}

	return result.Interfaces, response, nil
}

func (opts CreateOpts) request() (*createRequest, error) {
	if opts.Name == "" {
		return nil, invalid("the server name is required")
	}

	if opts.FlavorID == "" {
		return nil, invalid("the flavor ID is required")
	}

	networks, err := opts.networks()
	if err != nil {
		return nil, err
	}

	request := &createRequest{Server: createBody{
		Name:             opts.Name,
		FlavorRef:        opts.FlavorID,
		ImageRef:         opts.ImageID,
		AvailabilityZone: opts.AvailabilityZone,
		KeyName:          opts.KeyName,
		ConfigDrive:      opts.ConfigDrive,
		UserData:         opts.UserData,
		AdminPass:        opts.AdminPass,
		Metadata:         opts.Metadata,
		Tags:             opts.Tags,
		Networks:         networks,
		BlockDevices:     opts.BlockDevices,
	}}

	if opts.ServerGroupID != "" {
		request.SchedulerHints = &schedulerHints{Group: opts.ServerGroupID}
	}

	return request, nil
}

func (opts CreateOpts) networks() (any, error) {
	switch {
	case opts.NetworkMode != "" && len(opts.Networks) != 0:
		return nil, invalid("the networks and the network mode are mutually exclusive")
	case opts.NetworkMode == NetworkModeAuto || opts.NetworkMode == NetworkModeNone:
		return opts.NetworkMode, nil
	case opts.NetworkMode != "":
		return nil, invalid("the network mode must be auto or none")
	case len(opts.Networks) == 0:
		return nil, invalid("the networks or the network mode are required")
	}

	for _, network := range opts.Networks {
		if network.UUID == "" && network.Port == "" {
			return nil, invalid("a network needs the network UUID or the port UUID")
		}
	}

	return opts.Networks, nil
}

func serverPath(serverID string) (string, error) {
	if serverID == "" {
		return "", invalid("the server ID is required")
	}

	return basePath + "/" + url.PathEscape(serverID), nil
}

func tagPath(serverID, tag string) (string, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return "", err
	}

	if tag == "" {
		return "", invalid("the tag is required")
	}

	return path + "/tags/" + url.PathEscape(tag), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
