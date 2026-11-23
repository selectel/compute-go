package server

import (
	"context"
	"net/http"
	"testing"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/testhttp"
	"github.com/stretchr/testify/require"
)

const (
	testServerID       = "4f3c2b1a-0e9d-8c7b-6a59-483726150abc"
	testFlavorID       = "flavor-1"
	testImageID        = "image-1"
	testGroupID        = "group-1"
	testPortID         = "port-1"
	testVolumeID       = "volume-1"
	testNetworkID      = "net-1"
	testName           = "web"
	testOwner          = "team"
	testOwnerKey       = "owner"
	testPreemptibleTag = "preemptible"
	testAddress        = "10.0.0.5"

	apiNotFoundJSON = `{"itemNotFound": {"code": 404, "message": "Instance could not be found."}}`
	apiConflictJSON = `{"conflictingRequest": {"code": 409, "message": "Cannot 'confirmResize' instance."}}`

	apiServerJSON = `{"server": {
		"id": "4f3c2b1a-0e9d-8c7b-6a59-483726150abc",
		"name": "web",
		"status": "ERROR",
		"OS-EXT-STS:vm_state": "error",
		"OS-EXT-STS:task_state": null,
		"OS-EXT-STS:power_state": 0,
		"flavor": {
			"original_name": "SL1.1-1024-8", "vcpus": 1, "ram": 1024, "disk": 8, "ephemeral": 0, "swap": 0,
			"extra_specs": {"hw:cpu_policy": "shared"}
		},
		"image": {"id": "image-1", "links": []},
		"OS-EXT-AZ:availability_zone": "ru-9a",
		"key_name": "deploy",
		"config_drive": "True",
		"addresses": {"net": [
			{"addr": "10.0.0.5", "version": 4, "OS-EXT-IPS:type": "fixed", "OS-EXT-IPS-MAC:mac_addr": "fa:16:3e:00:00:01"}
		]},
		"accessIPv4": "10.0.0.5",
		"accessIPv6": "",
		"metadata": {"owner": "team"},
		"tags": ["preemptible", "web"],
		"server_groups": ["group-1"],
		"os-extended-volumes:volumes_attached": [{"id": "volume-1", "delete_on_termination": true}],
		"preemptible_forbidden_until": "2026-11-20T10:00:00+00:00",
		"fault": {"code": 500, "message": "No valid host was found.", "created": "2026-11-16T10:00:00Z"},
		"user_id": "user-1",
		"tenant_id": "project-1",
		"created": "2026-11-16T09:00:00Z",
		"updated": "2026-11-16T10:00:00Z"
	}}`
)

func serverURL(suffix string) string {
	return testhttp.BasePath + "/servers/" + testServerID + suffix
}

func answer(status int, body string) *testhttp.Client {
	return &testhttp.Client{Answers: []testhttp.Answer{{Status: status, Body: body}}}
}

func TestServer_Create(t *testing.T) {
	t.Run("SendsAllArguments", func(t *testing.T) {
		httpClient := answer(http.StatusAccepted, `{"server": {"id": "`+testServerID+`", "adminPass": "generated"}}`)

		created, response, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			Name:             testName,
			FlavorID:         testFlavorID,
			AvailabilityZone: "ru-9a",
			KeyName:          "deploy",
			ConfigDrive:      true,
			UserData:         "I2Nsb3VkLWNvbmZpZw==",
			AdminPass:        "secret",
			Metadata:         map[string]string{testOwnerKey: testOwner},
			Tags:             []string{testPreemptibleTag, testName},
			Networks:         []Network{{Port: testPortID}, {UUID: testNetworkID, FixedIP: testAddress}},
			BlockDevices: []BlockDevice{
				{
					SourceType: "image", DestinationType: "volume", UUID: testImageID, BootIndex: 0,
					VolumeSize: 10, DeleteOnTermination: true, VolumeType: "fast.ru-9a",
				},
				{SourceType: "blank", DestinationType: "volume", BootIndex: -1, VolumeSize: 5},
			},
			ServerGroupID: testGroupID,
		})
		require.NoError(t, err)
		require.Equal(t, testServerID, created.ID)
		require.Equal(t, http.StatusAccepted, response.StatusCode)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, testhttp.BasePath+"/servers", request.URL.Path)
		require.Equal(t, "compute "+v1.Microversion, request.Header.Get("OpenStack-API-Version"))
		require.JSONEq(t, `{
			"server": {
				"name": "web",
				"flavorRef": "flavor-1",
				"availability_zone": "ru-9a",
				"key_name": "deploy",
				"config_drive": true,
				"user_data": "I2Nsb3VkLWNvbmZpZw==",
				"adminPass": "secret",
				"metadata": {"owner": "team"},
				"tags": ["preemptible", "web"],
				"networks": [{"port": "port-1"}, {"uuid": "net-1", "fixed_ip": "10.0.0.5"}],
				"block_device_mapping_v2": [
					{"source_type": "image", "destination_type": "volume", "uuid": "image-1", "boot_index": 0,
					 "volume_size": 10, "delete_on_termination": true, "volume_type": "fast.ru-9a"},
					{"source_type": "blank", "destination_type": "volume", "boot_index": -1,
					 "volume_size": 5, "delete_on_termination": false}
				]
			},
			"os:scheduler_hints": {"group": "group-1"}
		}`, httpClient.LastBody(t))
	})

	t.Run("SendsImageAndNetworkMode", func(t *testing.T) {
		httpClient := answer(http.StatusAccepted, `{"server": {"id": "`+testServerID+`"}}`)

		_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
			Name: testName, FlavorID: testFlavorID, ImageID: testImageID, NetworkMode: NetworkModeNone,
		})
		require.NoError(t, err)
		require.JSONEq(
			t,
			`{"server": {"name": "web", "flavorRef": "flavor-1", "imageRef": "image-1", "networks": "none"}}`,
			httpClient.LastBody(t),
		)
	})

	t.Run("RejectsInvalidOptionsWithoutRequest", func(t *testing.T) {
		for name, opts := range map[string]CreateOpts{
			"NoName":           {FlavorID: testFlavorID, NetworkMode: NetworkModeNone},
			"NoFlavor":         {Name: testName, NetworkMode: NetworkModeNone},
			"NoNetworks":       {Name: testName, FlavorID: testFlavorID},
			"BothNetworkForms": {Name: testName, FlavorID: testFlavorID, NetworkMode: NetworkModeAuto, Networks: []Network{{UUID: "n"}}},
			"UnknownMode":      {Name: testName, FlavorID: testFlavorID, NetworkMode: "public"},
			"EmptyNetwork":     {Name: testName, FlavorID: testFlavorID, Networks: []Network{{FixedIP: testAddress}}},
		} {
			httpClient := answer(http.StatusAccepted, `{}`)

			_, _, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), opts)

			require.True(t, v1.IsKind(err, v1.KindInvalidRequest), name)
			require.Empty(t, httpClient.Requests, name)
		}
	})

	t.Run("ReportsFaultsByKind", func(t *testing.T) {
		for status, kind := range map[int]v1.Kind{
			http.StatusForbidden:           v1.KindForbidden,
			http.StatusConflict:            v1.KindConflict,
			http.StatusInternalServerError: v1.KindServerError,
		} {
			httpClient := answer(status, `{"fault": {"code": 0, "message": "Quota exceeded for instances."}}`)

			_, response, err := Create(t.Context(), testhttp.NewSDKClient(httpClient), CreateOpts{
				Name: testName, FlavorID: testFlavorID, NetworkMode: NetworkModeNone,
			})

			require.True(t, v1.IsKind(err, kind))
			require.Equal(t, status, response.StatusCode)
			require.Len(t, httpClient.Requests, 1)
		}
	})
}

func TestServer_Get(t *testing.T) {
	t.Run("DecodesTheServer", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiServerJSON)

		view, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
		require.NoError(t, err)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodGet, request.Method)
		require.Equal(t, serverURL(""), request.URL.Path)

		require.Equal(t, "ERROR", view.Status)
		require.Equal(t, "error", view.VMState)
		require.Empty(t, view.TaskState)
		require.Equal(t, "SL1.1-1024-8", view.Flavor.OriginalName)
		require.Equal(t, 1024, view.Flavor.RAM)
		require.Equal(t, testImageID, view.Image.ID)
		require.Equal(t, "ru-9a", view.AvailabilityZone)
		require.Equal(t, "deploy", view.KeyName)
		require.True(t, bool(view.ConfigDrive))
		require.Equal(t, "fa:16:3e:00:00:01", view.Addresses["net"][0].MAC)
		require.Equal(t, 4, view.Addresses["net"][0].Version)
		require.Equal(t, map[string]string{testOwnerKey: testOwner}, view.Metadata)
		require.Equal(t, []string{testPreemptibleTag, testName}, view.Tags)
		require.Equal(t, []string{testGroupID}, view.ServerGroups)
		require.Equal(t, []AttachedVolume{{ID: testVolumeID, DeleteOnTermination: true}}, view.VolumesAttached)
		require.Equal(t, ForbiddenUntil{Reported: true, Time: "2026-11-20T10:00:00+00:00"}, view.PreemptibleForbiddenUntil)
		require.Equal(t, &Fault{Code: 500, Message: "No valid host was found."}, view.Fault)
		require.Equal(t, 2026, view.Created.Year())
	})

	t.Run("DecodesAVolumeBackedServer", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"server": {
			"id": "`+testServerID+`", "status": "ACTIVE", "image": "", "config_drive": "",
			"preemptible_forbidden_until": null, "fault": null, "server_groups": [], "tags": []
		}}`)

		view, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
		require.NoError(t, err)
		require.Empty(t, view.Image.ID)
		require.False(t, bool(view.ConfigDrive))
		require.Equal(t, ForbiddenUntil{Reported: true}, view.PreemptibleForbiddenUntil,
			"null means that the pool sets a period that does not apply yet")
		require.Nil(t, view.Fault)
		require.Empty(t, view.ServerGroups)
	})

	t.Run("DecodesAPoolWithoutAForbidPeriod", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"server": {"id": "`+testServerID+`", "status": "ACTIVE"}}`)

		view, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
		require.NoError(t, err)
		require.Equal(t, ForbiddenUntil{}, view.PreemptibleForbiddenUntil)
	})

	t.Run("DistinguishesNotFoundFromConflict", func(t *testing.T) {
		_, _, err := Get(t.Context(), testhttp.NewSDKClient(answer(http.StatusNotFound, apiNotFoundJSON)), testServerID)
		require.True(t, v1.IsKind(err, v1.KindNotFound))
		require.False(t, v1.IsKind(err, v1.KindConflict))

		_, err = ConfirmResize(t.Context(), testhttp.NewSDKClient(answer(http.StatusConflict, apiConflictJSON)), testServerID)
		require.True(t, v1.IsKind(err, v1.KindConflict))
		require.False(t, v1.IsKind(err, v1.KindNotFound))
	})

	t.Run("RequiresTheServerID", func(t *testing.T) {
		httpClient := answer(http.StatusOK, apiServerJSON)

		_, _, err := Get(t.Context(), testhttp.NewSDKClient(httpClient), "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestServer_ListDetail(t *testing.T) {
	httpClient := &testhttp.Client{Answers: []testhttp.Answer{
		{Status: http.StatusOK, Body: `{"servers": [{"id": "a"}], "servers_links": [
			{"rel": "next", "href": "` + testhttp.Endpoint + `/servers/detail?marker=a"}]}`},
		{Status: http.StatusOK, Body: `{"servers": [{"id": "b"}]}`},
	}}

	servers, err := ListDetail(t.Context(), testhttp.NewSDKClient(httpClient), ListOpts{Name: "^tf-acc", Limit: 1})
	require.NoError(t, err)
	require.Len(t, servers, 2)
	require.Equal(t, testhttp.BasePath+"/servers/detail", httpClient.Requests[0].URL.Path)
	require.Equal(t, "^tf-acc", httpClient.Requests[0].URL.Query().Get("name"))
	require.Equal(t, "1", httpClient.Requests[0].URL.Query().Get("limit"))
	require.Equal(t, "a", httpClient.Requests[1].URL.Query().Get("marker"))
}

func TestServer_Update(t *testing.T) {
	name, userData := "web-2", "I2Nsb3VkLWNvbmZpZw=="

	for _, test := range []struct {
		name string
		opts UpdateOpts
		body string
	}{
		{name: "Name", opts: UpdateOpts{Name: &name}, body: `{"server": {"name": "web-2"}}`},
		{
			name: "NameAndUserData", opts: UpdateOpts{Name: &name, UserData: &userData},
			body: `{"server": {"name": "web-2", "user_data": "I2Nsb3VkLWNvbmZpZw=="}}`,
		},
		{name: "ClearUserData", opts: UpdateOpts{ClearUserData: true}, body: `{"server": {"user_data": null}}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			httpClient := answer(http.StatusOK, apiServerJSON)

			view, _, err := Update(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, test.opts)
			require.NoError(t, err)
			require.Equal(t, testServerID, view.ID)

			request := httpClient.LastRequest(t)
			require.Equal(t, http.MethodPut, request.Method)
			require.Equal(t, serverURL(""), request.URL.Path)
			require.JSONEq(t, test.body, httpClient.LastBody(t))
		})
	}

	t.Run("RejectsEmptyAndContradictoryOptions", func(t *testing.T) {
		for _, opts := range []UpdateOpts{{}, {UserData: &userData, ClearUserData: true}} {
			httpClient := answer(http.StatusOK, apiServerJSON)

			_, _, err := Update(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, opts)
			require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
			require.Empty(t, httpClient.Requests)
		}
	})
}

func TestServer_Delete(t *testing.T) {
	httpClient := answer(http.StatusNoContent, "")

	_, err := Delete(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
	require.NoError(t, err)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodDelete, request.Method)
	require.Equal(t, serverURL(""), request.URL.Path)
}

func TestServer_Actions(t *testing.T) {
	type call func(*testing.T, *v1.Client) (*v1.Response, error)

	simple := func(operation func(context.Context, *v1.Client, string) (*v1.Response, error)) call {
		return func(t *testing.T, client *v1.Client) (*v1.Response, error) {
			return operation(t.Context(), client, testServerID)
		}
	}

	for _, test := range []struct {
		name   string
		call   call
		status int
		body   string
	}{
		{name: "Start", call: simple(Start), status: http.StatusAccepted, body: `{"os-start": null}`},
		{name: "Stop", call: simple(Stop), status: http.StatusAccepted, body: `{"os-stop": null}`},
		{name: "Pause", call: simple(Pause), status: http.StatusAccepted, body: `{"pause": null}`},
		{name: "Unpause", call: simple(Unpause), status: http.StatusAccepted, body: `{"unpause": null}`},
		{name: "Shelve", call: simple(Shelve), status: http.StatusAccepted, body: `{"shelve": null}`},
		{name: "Unshelve", call: simple(Unshelve), status: http.StatusAccepted, body: `{"unshelve": null}`},
		{
			name: "ConfirmResize", call: simple(ConfirmResize), status: http.StatusNoContent,
			body: `{"confirmResize": null}`,
		},
		{
			name: "LeaveServerGroup", call: simple(LeaveServerGroup), status: http.StatusOK,
			body: `{"leaveServerGroup": {}}`,
		},
		{
			name: "Resize", status: http.StatusAccepted, body: `{"resize": {"flavorRef": "flavor-1"}}`,
			call: func(t *testing.T, client *v1.Client) (*v1.Response, error) {
				return Resize(t.Context(), client, testServerID, testFlavorID)
			},
		},
		{
			name: "Rebuild", status: http.StatusAccepted, body: `{"rebuild": {"imageRef": "image-1"}}`,
			call: func(t *testing.T, client *v1.Client) (*v1.Response, error) {
				return Rebuild(t.Context(), client, testServerID, RebuildOpts{ImageID: testImageID})
			},
		},
		{
			name: "ChangePassword", status: http.StatusAccepted, body: `{"changePassword": {"adminPass": "secret"}}`,
			call: func(t *testing.T, client *v1.Client) (*v1.Response, error) {
				return ChangePassword(t.Context(), client, testServerID, "secret")
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			httpClient := answer(test.status, apiServerJSON)

			_, err := test.call(t, testhttp.NewSDKClient(httpClient))
			require.NoError(t, err)

			request := httpClient.LastRequest(t)
			require.Equal(t, http.MethodPost, request.Method)
			require.Equal(t, serverURL("/action"), request.URL.Path)
			require.JSONEq(t, test.body, httpClient.LastBody(t))

			httpClient = answer(http.StatusConflict, apiConflictJSON)

			_, err = test.call(t, testhttp.NewSDKClient(httpClient))
			require.True(t, v1.IsKind(err, v1.KindConflict))
		})
	}

	t.Run("RejectsEmptyArgumentsWithoutRequest", func(t *testing.T) {
		httpClient := answer(http.StatusAccepted, "")
		client := testhttp.NewSDKClient(httpClient)

		_, err := Resize(t.Context(), client, testServerID, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, err = Rebuild(t.Context(), client, testServerID, RebuildOpts{})
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, err = ChangePassword(t.Context(), client, testServerID, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, err = Start(t.Context(), client, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		require.Empty(t, httpClient.Requests)
	})
}

func TestServer_Metadata(t *testing.T) {
	t.Run("SetMergesKeys", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"metadata": {"owner": "team", "env": "prod"}}`)

		metadata, _, err := SetMetadata(
			t.Context(), testhttp.NewSDKClient(httpClient), testServerID, map[string]string{"env": "prod"},
		)
		require.NoError(t, err)
		require.Equal(t, map[string]string{testOwnerKey: testOwner, "env": "prod"}, metadata)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodPost, request.Method)
		require.Equal(t, serverURL("/metadata"), request.URL.Path)
		require.JSONEq(t, `{"metadata": {"env": "prod"}}`, httpClient.LastBody(t))
	})

	t.Run("DeleteRemovesOneKey", func(t *testing.T) {
		httpClient := answer(http.StatusNoContent, "")

		_, err := DeleteMetadatum(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, "a key")
		require.NoError(t, err)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodDelete, request.Method)
		require.Equal(t, serverURL("/metadata/a key"), request.URL.Path)
	})

	t.Run("RejectsEmptyArgumentsWithoutRequest", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{}`)
		client := testhttp.NewSDKClient(httpClient)

		_, _, err := SetMetadata(t.Context(), client, testServerID, nil)
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))

		_, err = DeleteMetadatum(t.Context(), client, testServerID, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestServer_Tags(t *testing.T) {
	t.Run("AddAcceptsCreatedAndAlreadyPresent", func(t *testing.T) {
		for _, status := range []int{http.StatusCreated, http.StatusNoContent} {
			httpClient := answer(status, "")

			_, err := AddTag(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testPreemptibleTag)
			require.NoError(t, err)

			request := httpClient.LastRequest(t)
			require.Equal(t, http.MethodPut, request.Method)
			require.Equal(t, serverURL("/tags/preemptible"), request.URL.Path)
			require.Empty(t, httpClient.LastBody(t))
		}
	})

	t.Run("AddReportsForbidden", func(t *testing.T) {
		httpClient := answer(
			http.StatusForbidden,
			`{"forbidden": {"code": 403, "message": "Setting instance to preemptible is forbidden until 2026-11-20."}}`,
		)

		_, err := AddTag(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testPreemptibleTag)
		require.True(t, v1.IsKind(err, v1.KindForbidden))
		require.ErrorContains(t, err, "forbidden until")
	})

	t.Run("DeleteRemovesOneTag", func(t *testing.T) {
		httpClient := answer(http.StatusNoContent, "")

		_, err := DeleteTag(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, testName)
		require.NoError(t, err)

		request := httpClient.LastRequest(t)
		require.Equal(t, http.MethodDelete, request.Method)
		require.Equal(t, serverURL("/tags/web"), request.URL.Path)
	})

	t.Run("List", func(t *testing.T) {
		httpClient := answer(http.StatusOK, `{"tags": ["preemptible", "web"]}`)

		tags, _, err := ListTags(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
		require.NoError(t, err)
		require.Equal(t, []string{testPreemptibleTag, testName}, tags)
		require.Equal(t, serverURL("/tags"), httpClient.LastRequest(t).URL.Path)
	})

	t.Run("RequiresTheTag", func(t *testing.T) {
		httpClient := answer(http.StatusNoContent, "")

		_, err := AddTag(t.Context(), testhttp.NewSDKClient(httpClient), testServerID, "")
		require.True(t, v1.IsKind(err, v1.KindInvalidRequest))
		require.Empty(t, httpClient.Requests)
	})
}

func TestServer_ListInterfaces(t *testing.T) {
	httpClient := answer(http.StatusOK, `{"interfaceAttachments": [{
		"port_id": "port-1", "net_id": "net-1", "mac_addr": "fa:16:3e:00:00:01", "port_state": "ACTIVE",
		"fixed_ips": [{"ip_address": "10.0.0.5", "subnet_id": "subnet-1"}]
	}]}`)

	interfaces, _, err := ListInterfaces(t.Context(), testhttp.NewSDKClient(httpClient), testServerID)
	require.NoError(t, err)
	require.Equal(t, []Interface{{
		PortID: testPortID, NetworkID: testNetworkID, MAC: "fa:16:3e:00:00:01", PortState: "ACTIVE",
		FixedIPs: []FixedIP{{Address: testAddress, SubnetID: "subnet-1"}},
	}}, interfaces)

	request := httpClient.LastRequest(t)
	require.Equal(t, http.MethodGet, request.Method)
	require.Equal(t, serverURL("/os-interface"), request.URL.Path)
}
