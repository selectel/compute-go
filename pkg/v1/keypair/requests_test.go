package keypair

import (
	"net/http"
	"testing"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

func answer(status int, body string) *testhttp.Client {
	return &testhttp.Client{Answers: []testhttp.Answer{{Status: status, Body: body}}}
}

const apiForbiddenJSON = `{"forbidden": {"code": 403, "message": "Policy does not allow this to be performed."}}`

const (
	testKeyName   = "deploy key"
	testPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample deploy"
	//nolint:gosec // G101: this is a fake public key used by the tests.
	apiKeyPairJSON = `{"keypair": {
		"name": "deploy key", "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample deploy",
		"fingerprint": "SHA256:abc", "type": "ssh", "user_id": "user-1"
	}}`
)

func TestKeyPair(t *testing.T) {
	t.Run("ImportUploadsThePublicKey", func(t *testing.T) {
		httpClient := answer(http.StatusCreated, apiKeyPairJSON)

		view, _, err := Import(t.Context(), testhttp.NewSDKClient(httpClient), testKeyName, testPublicKey)
		require.NoError(t, err)
		require.Equal(t, "SHA256:abc", view.Fingerprint)
		require.Equal(t, "user-1", view.UserID)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, testhttp.BasePath+"/os-keypairs", request.URL.Path)
		require.JSONEq(
			t, `{"keypair": {"name": "deploy key", "public_key": "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample deploy"}}`,
			httpClient.LastBody(t),
		)
	})

	t.Run("GetAndDeleteEscapeTheName", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiKeyPairJSON)

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testKeyName)
		require.NoError(t, err)
		require.Equal(t, testhttp.BasePath+"/os-keypairs/deploy key", httpClient.LastRequest(t).URL.Path)
		require.Equal(t, testhttp.BasePath+"/os-keypairs/deploy%20key", httpClient.LastRequest(t).URL.EscapedPath())

		httpClient = answer(http.StatusNoContent, "")

		_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), testKeyName)
		require.NoError(t, err)
		require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
	})

	t.Run("ListUnwrapsItems", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"keypairs": [{"keypair": {"name": "a"}}, {"keypair": {"name": "b"}}]}`)

		keys, err := List(t.Context(), testhttp.NewSDKClient(httpClient))
		require.NoError(t, err)
		require.Equal(t, []View{{Name: "a"}, {Name: "b"}}, keys)
	})

	t.Run("Errors", func(t *testing.T) {
		_, _, err := Import(
			t.Context(), testhttp.NewSDKClient(answer(http.StatusConflict, `{"conflictingRequest": {"code": 409, "message": "Key pair 'deploy key' already exists."}}`)),
			testKeyName, testPublicKey,
		)
		require.True(t, v1.IsKind(err, v1.KindConflict))

		_, _, err = Get(t.Context(), testhttp.NewSDKClient(answer(http.StatusForbidden, apiForbiddenJSON)), testKeyName)
		require.True(t, v1.IsKind(err, v1.KindForbidden))

		httpClient := answer(http.StatusCreated, apiKeyPairJSON)

		_, _, err = Import(t.Context(), testhttp.NewSDKClient(httpClient), testKeyName, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}
