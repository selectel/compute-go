package volumeattach

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const envelope = "volumeAttachment"

// Create asks the cloud to attach the volume. The attachment completes asynchronously.
func Create(ctx context.Context, client *v1.Client, serverID string, opts CreateOpts) (*View, *v1.Response, error) {
	path, err := attachmentsPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	if opts.VolumeID == "" {
		return nil, nil, invalid("the volume ID is required")
	}

	request := createRequest{VolumeAttachment: createBody(opts)}
	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, path, http.StatusOK, request, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.VolumeAttachment, response, nil
}

func Get(ctx context.Context, client *v1.Client, serverID, volumeID string) (*View, *v1.Response, error) {
	path, err := attachmentPath(serverID, volumeID)
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

	return &result.VolumeAttachment, response, nil
}

func List(ctx context.Context, client *v1.Client, serverID string) ([]View, *v1.Response, error) {
	path, err := attachmentsPath(serverID)
	if err != nil {
		return nil, nil, err
	}

	result := &listEnvelope{}

	response, err := transport.DoRequest(ctx, client, http.MethodGet, path, http.StatusOK, nil, result)
	if err != nil {
		return nil, response, err
	}

	return result.VolumeAttachments, response, nil
}

// Delete asks the cloud to detach the volume. The API answers with an invalid request while the
// server or the volume is in a transitional state.
func Delete(ctx context.Context, client *v1.Client, serverID, volumeID string) (*v1.Response, error) {
	path, err := attachmentPath(serverID, volumeID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusAccepted, nil, nil)
}

func attachmentsPath(serverID string) (string, error) {
	if serverID == "" {
		return "", invalid("the server ID is required")
	}

	return "servers/" + url.PathEscape(serverID) + "/os-volume_attachments", nil
}

func attachmentPath(serverID, volumeID string) (string, error) {
	path, err := attachmentsPath(serverID)
	if err != nil {
		return "", err
	}

	if volumeID == "" {
		return "", invalid("the volume ID is required")
	}

	return path + "/" + url.PathEscape(volumeID), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
