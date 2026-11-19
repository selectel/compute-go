package flavor

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
	testFlavorID   = "custom-1"
	testFlavorName = "custom"
	apiFlavorJSON  = `{"flavor": {
		"id": "custom-1", "name": "custom", "description": "a flavor", "vcpus": 2, "ram": 4096, "disk": 0,
		"OS-FLV-EXT-DATA:ephemeral": 0, "swap": 0, "rxtx_factor": "", "os-flavor-access:is_public": false,
		"OS-FLV-DISABLED:disabled": false, "extra_specs": {"hw:cpu_policy": "dedicated"},
		"availability_zones": ["ru-9a", "ru-9b"]
	}}`
)

func TestFlavor_Create(t *testing.T) {
	t.Run("SendsLineParameters", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiFlavorJSON)
		enabled := true

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			ID: testFlavorID, Name: testFlavorName, Description: "a flavor", VCPUs: 2, RAM: 4096,
			GPU: "a100:1", DedicatedCPUs: &enabled, NetworkQoS: 2, AvailabilityZone: "ru-9a",
			HighFreq: &enabled, VCPUsVendor: "amd",
		})
		require.NoError(t, err)
		require.Equal(t, testFlavorID, view.ID)
		require.Equal(t, []string{"ru-9a", "ru-9b"}, view.AvailabilityZones)
		require.Equal(t, map[string]string{"hw:cpu_policy": "dedicated"}, view.ExtraSpecs)
		require.Zero(t, view.RxTxFactor)
		require.False(t, view.IsPublic)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, testhttp.BasePath+"/flavors", request.URL.Path)
		require.JSONEq(t, `{"flavor": {
			"id": "custom-1", "name": "custom", "description": "a flavor", "vcpus": 2, "ram": 4096, "disk": 0,
			"os-flavor-access:is_public": false, "gpu": "a100:1", "dedicated_cpus": true, "network_qos": 2,
			"availability_zone": "ru-9a", "highfreq": true, "vcpus_vendor": "amd"
		}}`, httpClient.LastBody(t))
	})

	t.Run("OmitsUnsetParameters", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiFlavorJSON)

		_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{Name: testFlavorName, VCPUs: 1, RAM: 512, Disk: 5})
		require.NoError(t, err)
		require.JSONEq(
			t,
			`{"flavor": {"name": "custom", "vcpus": 1, "ram": 512, "disk": 5, "os-flavor-access:is_public": false}}`,
			httpClient.LastBody(t),
		)
	})

	t.Run("RejectsInvalidSizesWithoutRequest", func(t *testing.T) {
		for _, opts := range []CreateOpts{
			{VCPUs: 1, RAM: 512}, {Name: testFlavorName, RAM: 512}, {Name: testFlavorName, VCPUs: 1}, {Name: testFlavorName, VCPUs: 1, RAM: 512, Disk: -1},
		} {
			httpClient := answer(http.StatusOK, apiFlavorJSON)

			_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), opts)
			require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
			require.Empty(t, httpClient.Requests)
		}
	})

	t.Run("ReportsFaultsByKind", func(t *testing.T) {
		_, _, err := Create(
			t.Context(), testhttp.NewSDKClient(answer(http.StatusConflict, `{"conflictingRequest": {"code": 409, "message": "Flavor with ID custom-1 already exists."}}`)),
			CreateOpts{Name: testFlavorName, VCPUs: 1, RAM: 512},
		)
		require.True(t, v1.IsKind(err, v1.KindConflict))
		require.ErrorContains(t, err, "already exists")
	})
}

func TestFlavor_GetAndDelete(t *testing.T) {
	httpClient := answer(http.StatusOK, `{"flavor": {"id": "1311", "name": "SL1.1-1024-8", "rxtx_factor": 1.0, "os-flavor-access:is_public": true}}`)

	view, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), "1311")
	require.NoError(t, err)
	require.True(t, view.IsPublic)
	require.InDelta(t, 1.0, float64(view.RxTxFactor), 0)
	require.Equal(t, http.MethodGet, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/flavors/1311", httpClient.LastRequest(t).URL.Path)

	httpClient = answer(http.StatusAccepted, "")
	_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), testFlavorID)
	require.NoError(t, err)
	require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
	require.Equal(t, testhttp.BasePath+"/flavors/"+testFlavorID, httpClient.LastRequest(t).URL.Path)

	_, err = Delete(t.Context(), testhttp.NewSDKClient(answer(http.StatusForbidden, apiForbiddenJSON)), testFlavorID)
	require.True(t, v1.IsKind(err, v1.KindForbidden))

	_, _, err = Get(t.Context(), testhttp.NewSDKClient(httpClient), "")
	require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
}

func TestFlavor_ListDetail(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{
		{Status: http.StatusOK, Body: `{"flavors": [{"id": "a"}], "flavors_links": [
			{"rel": "next", "href": "` + testhttp.Endpoint + `/flavors/detail?marker=a"}]}`},
		{Status: http.StatusOK, Body: `{"flavors": [{"id": "b"}]}`},
	}}

	flavors, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{IsPublic: "none"})
	require.NoError(t, err)
	require.Len(t, flavors, 2)
	require.Equal(t, testhttp.BasePath+"/flavors/detail", httpClient.Requests[0].URL.Path)
	require.Equal(t, "none", httpClient.Requests[0].URL.Query().Get("is_public"))
	require.Equal(t, "a", httpClient.Requests[1].URL.Query().Get("marker"))

	httpClient = &testhttp.Client{Answers: []testhttp.Answer{
		{Status: http.StatusOK, Body: `{"flavors": [{"id": "a"}], "flavors_links": [
			{"rel": "next", "href": "` + testhttp.Endpoint + `/flavors/detail?marker=a"}]}`},
		{Status: http.StatusInternalServerError, Body: `{}`},
	}}

	flavors, err = ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{})
	require.True(t, v1.IsKind(err, v1.KindIncompleteList))
	require.Empty(t, flavors)
}
