package servergroup

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
	testGroupID   = "group-1"
	testGroupName = "web"
	apiGroupJSON  = `{"server_group": {
		"id": "group-1", "name": "web", "policy": "anti-affinity", "rules": {"max_server_per_host": 2},
		"members": ["server-1"], "project_id": "project-1", "user_id": "user-1"
	}}`
)

func TestServerGroup(t *testing.T) {
	t.Run("CreateSendsPolicyAndRule", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiGroupJSON)

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			Name: testGroupName, Policy: "anti-affinity", MaxServerPerHost: 2,
		})
		require.NoError(t, err)
		require.Equal(t, 2, view.Rules.MaxServerPerHost)
		require.Equal(t, []string{"server-1"}, view.Members)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, testhttp.BasePath+"/os-server-groups", request.URL.Path)
		require.JSONEq(
			t, `{"server_group": {"name": "web", "policy": "anti-affinity", "rules": {"max_server_per_host": 2}}}`,
			httpClient.LastBody(t),
		)
	})

	t.Run("CreateOmitsUnsetRule", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"server_group": {"id": "group-1", "policy": "soft-anti-affinity", "rules": {}}}`)

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{Name: testGroupName, Policy: "soft-anti-affinity"})
		require.NoError(t, err)
		require.Zero(t, view.Rules.MaxServerPerHost)
		require.JSONEq(t, `{"server_group": {"name": "web", "policy": "soft-anti-affinity"}}`, httpClient.LastBody(t))
	})

	t.Run("GetListDelete", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiGroupJSON)

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testGroupID)
		require.NoError(t, err)
		require.Equal(t, testhttp.BasePath+"/os-server-groups/"+testGroupID, httpClient.LastRequest(t).URL.Path)

		httpClient = answer(http.StatusOK, `{"server_groups": [{"id": "group-1"}, {"id": "group-2"}]}`)

		groups, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient))
		require.NoError(t, err)
		require.Len(t, groups, 2)
		require.Equal(t, testhttp.BasePath+"/os-server-groups", httpClient.LastRequest(t).URL.Path)

		httpClient = answer(http.StatusNoContent, "")

		_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), testGroupID)
		require.NoError(t, err)
		require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
	})

	t.Run("Errors", func(t *testing.T) {
		_, _, err := Get(t.Context(), testhttp.NewSDKClient(answer(http.StatusForbidden, apiForbiddenJSON)), testGroupID)
		require.True(t, v1.IsKind(err, v1.KindForbidden))

		httpClient := answer(http.StatusOK, apiGroupJSON)

		_, _, err = Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{Name: testGroupName})
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}
