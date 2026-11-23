package keypair

import (
	"context"
	"net/http"
	"net/url"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/pagination"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const (
	basePath = "os-keypairs"
	envelope = "keypair"
)

// Import uploads a public key. The API does not generate key pairs at this microversion.
func Import(ctx context.Context, client *v1.Client, name, publicKey string) (*View, *v1.Response, error) {
	if name == "" || publicKey == "" {
		return nil, nil, invalid("the key pair name and the public key are required")
	}

	request := createRequest{KeyPair: createBody{Name: name, PublicKey: publicKey}}
	result := &viewEnvelope{}

	response, err := transport.DoRequest(
		ctx, client, http.MethodPost, basePath, http.StatusCreated, request, result,
		transport.WithResponseEnvelope(envelope),
	)
	if err != nil {
		return nil, response, err
	}

	return &result.KeyPair, response, nil
}

func Get(ctx context.Context, client *v1.Client, name string) (*View, *v1.Response, error) {
	path, err := keyPairPath(name)
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

	return &result.KeyPair, response, nil
}

// List returns all key pairs of the user and never returns a partial result.
func List(ctx context.Context, client *v1.Client) ([]View, error) {
	return pagination.ReadAll[View](ctx, client, basePath, "keypairs", nil, 0, func() *pageEnvelope {
		return &pageEnvelope{}
	})
}

func Delete(ctx context.Context, client *v1.Client, name string) (*v1.Response, error) {
	path, err := keyPairPath(name)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodDelete, path, http.StatusNoContent, nil, nil)
}

func keyPairPath(name string) (string, error) {
	if name == "" {
		return "", invalid("the key pair name is required")
	}

	return basePath + "/" + url.PathEscape(name), nil
}

func invalid(message string) *v1.Error {
	return &v1.Error{Kind: v1.KindInvalidRequest, Message: message}
}
