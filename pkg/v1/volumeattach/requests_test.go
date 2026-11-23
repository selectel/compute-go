package volumeattach

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
	testServerID      = "server-1"
	testVolumeID      = "volume-1"
	apiAttachmentJSON = `{"volumeAttachment": {
		"serverId": "server-1", "volumeId": "volume-1", "device": "/dev/sdb", "tag": "data",
		"delete_on_termination": false, "attachment_id": "attachment-1", "bdm_uuid": "bdm-1"
	}}`
)

func attachmentsURL(suffix string) string {
	return testhttp.BasePath + "/servers/" + testServerID + "/os-volume_attachments" + suffix
}

func TestVolumeAttach(t *testing.T) {
	t.Run("CreateSendsTheTag", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiAttachmentJSON)

		view, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, CreateOpts{VolumeID: testVolumeID, Tag: "data"})
		require.NoError(t, err)
		require.Equal(t, View{ServerID: testServerID, VolumeID: testVolumeID, Device: "/dev/sdb", Tag: "data"}, *view)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, attachmentsURL(""), request.URL.Path)
		require.JSONEq(t, `{"volumeAttachment": {"volumeId": "volume-1", "tag": "data"}}`, httpClient.LastBody(t))
	})

	t.Run("GetListDelete", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiAttachmentJSON)

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testVolumeID)
		require.NoError(t, err)
		require.Equal(t, attachmentsURL("/"+testVolumeID), httpClient.LastRequest(t).URL.Path)

		httpClient = answer(http.StatusOK, `{"volumeAttachments": [{"volumeId": "volume-1"}]}`)

		attachments, _, err := List(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
		require.NoError(t, err)
		require.Len(t, attachments, 1)
		require.Equal(t, attachmentsURL(""), httpClient.LastRequest(t).URL.Path)

		httpClient = answer(http.StatusAccepted, "")

		_, err = Delete(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testVolumeID)
		require.NoError(t, err)
		require.Equal(t, http.MethodDelete, httpClient.LastRequest(t).Method)
		require.Equal(t, attachmentsURL("/"+testVolumeID), httpClient.LastRequest(t).URL.Path)
	})

	t.Run("Errors", func(t *testing.T) {
		for status, kind := range map[int]v1.Kind{
			http.StatusBadRequest: v1.KindInvalidRequest,
			http.StatusNotFound:   v1.KindNotFound,
			http.StatusConflict:   v1.KindConflict,
			http.StatusForbidden:  v1.KindForbidden,
		} {
			_, err := Delete(t.Context(), testhttp.NewSDKClient(answer(status, apiForbiddenJSON)), testServerID, testVolumeID)
			require.True(t, v1.IsKind(err, kind))
		}

		httpClient := answer(http.StatusOK, apiAttachmentJSON)

		_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, CreateOpts{})
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, _, err = Get(t.Context(), testhttp.NewSDKClient(httpClient), "", testVolumeID)
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}
