package interfaceattach

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
	testServerID     = "server-1"
	testPortID       = "port-1"
	testNetworkID    = "net-1"
	testAddress      = "10.0.0.5"
	apiInterfaceJSON = `{"interfaceAttachment": {
		"port_id": "port-1", "net_id": "net-1", "mac_addr": "fa:16:3e:00:00:01", "port_state": "DOWN",
		"fixed_ips": [{"ip_address": "10.0.0.5", "subnet_id": "subnet-1"}]
	}}`
)

func interfacesURL(suffix string) string {
	return testhttp.BasePath + "/servers/" + testServerID + "/os-interface" + suffix
}

func TestInterfaceAttach(t *testing.T) {
	t.Run("CreateWithPort", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiInterfaceJSON)

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, CreateOpts{PortID: testPortID})
		require.NoError(t, err)
		require.Equal(t, testNetworkID, view.NetworkID)
		require.Equal(t, []FixedIP{{Address: testAddress, SubnetID: "subnet-1"}}, view.FixedIPs)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, interfacesURL(""), request.URL.Path)
		require.JSONEq(t, `{"interfaceAttachment": {"port_id": "port-1"}}`, httpClient.LastBody(t))
	})

	t.Run("CreateWithNetworkAndAddress", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiInterfaceJSON)

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, CreateOpts{NetworkID: testNetworkID, FixedIP: testAddress})
		require.NoError(t, err)
		require.Equal(t, testPortID, view.PortID)
		require.JSONEq(
			t, `{"interfaceAttachment": {"net_id": "net-1", "fixed_ips": [{"ip_address": "10.0.0.5"}]}}`,
			httpClient.LastBody(t),
		)
	})

	t.Run("CreateRejectsInvalidOptionsWithoutRequest", func(t *testing.T) {
		for _, opts := range []CreateOpts{
			{}, {PortID: testPortID, NetworkID: testNetworkID}, {PortID: testPortID, FixedIP: testAddress},
		} {
			httpClient := answer(http.StatusOK, apiInterfaceJSON)

			_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, opts)
			require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
			require.Empty(t, httpClient.Requests)
		}
	})

	t.Run("GetAndDelete", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiInterfaceJSON)

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testPortID)
		require.NoError(t, err)
		require.Equal(t, interfacesURL("/"+testPortID), httpClient.LastRequest(t).URL.Path)

		httpClient = answer(http.StatusAccepted, "")

		_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testPortID)
		require.NoError(t, err)
		require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
		require.Equal(t, interfacesURL("/"+testPortID), httpClient.LastRequest(t).URL.Path)

		_, err = Delete(t.Context(), testhttp.NewSDKClient(answer(http.StatusForbidden, apiForbiddenJSON)), testServerID, testPortID)
		require.True(t, v1.IsKind(err, v1.KindForbidden))

		_, err = Delete(t.Context(), testhttp.NewSDKClient(answer(http.StatusNotFound, apiForbiddenJSON)), testServerID, testPortID)
		require.True(t, v1.IsKind(err, v1.KindNotFound))
	})
}
