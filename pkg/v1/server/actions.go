package server

import (
	"context"
	"net/http"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

func Start(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"os-start": nil})
}

func Stop(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"os-stop": nil})
}

func Pause(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"pause": nil})
}

func Unpause(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"unpause": nil})
}

// Shelve stops the server and releases its host. The cloud offloads a shelved server at once.
func Shelve(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"shelve": nil})
}

func Unshelve(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusAccepted, action{"unshelve": nil})
}

// Resize starts a flavor change. The server then waits in VERIFY_RESIZE for a confirmation
// unless the cloud confirms the change itself.
func Resize(ctx context.Context, client *v1.Client, serverID, flavorID string) (*v1.Response, error) {
	if flavorID == "" {
		return nil, invalid("the flavor ID is required")
	}

	return doAction(ctx, client, serverID, http.StatusAccepted, action{"resize": resizeBody{FlavorRef: flavorID}})
}

func ConfirmResize(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusNoContent, action{"confirmResize": nil})
}

// Rebuild replaces the content of the local disk with the image and keeps the server ID.
func Rebuild(ctx context.Context, client *v1.Client, serverID string, opts RebuildOpts) (*v1.Response, error) {
	if opts.ImageID == "" {
		return nil, invalid("the image ID is required")
	}

	return doAction(ctx, client, serverID, http.StatusAccepted, action{"rebuild": rebuildBody{ImageRef: opts.ImageID}})
}

// ChangePassword needs a running server with a guest agent.
func ChangePassword(ctx context.Context, client *v1.Client, serverID, password string) (*v1.Response, error) {
	if password == "" {
		return nil, invalid("the password is required")
	}

	return doAction(
		ctx, client, serverID, http.StatusAccepted, action{"changePassword": changePasswordBody{AdminPass: password}},
	)
}

// LeaveServerGroup removes the server from its server group in place. The API answers with an
// invalid request when the server is not a member of a group.
func LeaveServerGroup(ctx context.Context, client *v1.Client, serverID string) (*v1.Response, error) {
	return doAction(ctx, client, serverID, http.StatusOK, action{"leaveServerGroup": struct{}{}})
}

func doAction(
	ctx context.Context, client *v1.Client, serverID string, expectedStatus int, body action,
) (*v1.Response, error) {
	path, err := serverPath(serverID)
	if err != nil {
		return nil, err
	}

	return transport.DoRequest(ctx, client, http.MethodPost, path+"/action", expectedStatus, body, nil)
}
