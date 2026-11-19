package availabilityzone

import (
	"context"
	"net/http"

	v1 "github.com/selectel/compute-go/pkg/v1"
	"github.com/selectel/compute-go/pkg/v1/internal/transport"
)

const basePath = "os-availability-zone"

// List returns the availability zones of the region that the project can see.
func List(ctx context.Context, client *v1.Client) ([]View, *v1.Response, error) {
	result := &listEnvelope{}

	response, err := transport.DoRequest(ctx, client, http.MethodGet, basePath, http.StatusOK, nil, result)
	if err != nil {
		return nil, response, err
	}

	return result.Zones, response, nil
}
