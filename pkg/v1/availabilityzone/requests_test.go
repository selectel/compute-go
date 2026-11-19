package availabilityzone

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

func TestAvailabilityZone_List(t *testing.T) {
	httpClient := answer(http.StatusOK, `{"availabilityZoneInfo": [
		{"zoneName": "ru-9a", "zoneState": {"available": true}, "hosts": null},
		{"zoneName": "ru-9b", "zoneState": {"available": false}, "hosts": null}
	]}`)

	zones, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient))
	require.NoError(t, err)
	require.Equal(t, []View{{Name: "ru-9a", State: State{Available: true}}, {Name: "ru-9b"}}, zones)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodGet, request.Method)
	require.Equal(t, testhttp.BasePath+"/os-availability-zone", request.URL.Path)

	_, _, err = List(t.Context(), testhttp.NewSDKClient(answer(http.StatusForbidden, apiForbiddenJSON)))
	require.True(t, v1.IsKind(err, v1.KindForbidden))
}
