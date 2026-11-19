package interfaceattach

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const envelope = "interfaceAttachment"

// Create attaches an existing port, or asks the cloud to create a port in the network.
func Create(ctx context.Context, client *v1.Client, serverID string, opts CreateOpts) (*View, *v1.Response, error) {
	path, err := interfacesPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	if (opts.PortID == "") == (opts.NetworkID == "") {
		return nil, nil, invalid("exactly one of the port ID and the network ID is required")
	}

	if opts.FixedIP != "" && opts.NetworkID == "" {
		return nil, nil, invalid("a fixed IP address needs the network ID")
	}

	body := createBody{PortID: opts.PortID, NetworkID: opts.NetworkID}
	if opts.FixedIP != "" {
		body.FixedIPs = []requestFixedIP{{Address: opts.FixedIP}}
	}

	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, path, http.StatusOK, createRequest{InterfaceAttachment: body}, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.InterfaceAttachment, response, nil
}

func Get(ctx context.Context, client *v1.Client, serverID, portID string) (*View, *v1.Response, error) {
	path, err := interfacePath(serverID, portID)
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

	return &result.InterfaceAttachment, response, nil
}

// Delete asks the cloud to detach the port. The detachment completes asynchronously.
func Delete(ctx context.Context, client *v1.Client, serverID, portID string) (*v1.Response, error) {
	path, err := interfacePath(serverID, portID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusAccepted, nil, nil)
}

func interfacesPath(serverID string) (string, error) {
	if serverID == "" {
		return "", invalid("the server ID is required")
	}

	return "servers/" + url.PathEscape(serverID) + "/os-interface", nil
}

func interfacePath(serverID, portID string) (string, error) {
	path, err := interfacesPath(serverID)
	if err != nil {
		return "", err
	}

	if portID == "" {
		return "", invalid("the port ID is required")
	}

	return path + "/" + url.PathEscape(portID), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
