package flavor

import (
	"encoding/json"

	"github.com/selectel/compute-go/pkg/v1/internal/pagination"
)

type View struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	VCPUs       int    `json:"vcpus"`

	// RAM is in MiB; Disk and Ephemeral are in GiB; Swap is in MiB.
	RAM       int `json:"ram"`
	Disk      int `json:"disk"`
	Ephemeral int `json:"OS-FLV-EXT-DATA:ephemeral"`
	Swap      int `json:"swap"`

	RxTxFactor RxTxFactor `json:"rxtx_factor"`
	IsPublic   bool       `json:"os-flavor-access:is_public"`
	Disabled   bool       `json:"OS-FLV-DISABLED:disabled"`

	// ExtraSpecs show the outcome of the line parameters given at creation.
	ExtraSpecs map[string]string `json:"extra_specs"`

	// AvailabilityZones are the zones where the flavor can be used.
	AvailabilityZones []string `json:"availability_zones"`
}

// RxTxFactor is reported by the API as a number or as an empty string.
type RxTxFactor float64

func (factor *RxTxFactor) UnmarshalJSON(raw []byte) error {
	var value float64
	if err := json.Unmarshal(raw, &value); err != nil {
		*factor = 0

		return nil //nolint:nilerr // The API reports an unset factor as an empty string.
	}

	*factor = RxTxFactor(value)

	return nil
}

// CreateOpts describe a flavor. Name, VCPUs and RAM are required.
type CreateOpts struct {
	// ID is optional; the cloud generates one when it is empty.
	ID          string
	Name        string
	Description string
	VCPUs       int
	RAM         int
	Disk        int
	IsPublic    bool

	// The line parameters select a product line of the cloud. The API does not return them.
	GPU              string
	DedicatedCPUs    *bool
	NetworkQoS       int
	AvailabilityZone string
	HighFreq         *bool
	UltraHighFreq    *bool
	VCPUsVendor      string
}

// ListOpts are server-side filters.
type ListOpts struct {
	// IsPublic is "true", "false" or "none". The API applies it only for tokens that may see
	// flavors of every kind; other tokens get the flavors visible to their project.
	IsPublic string

	// Limit is the page size. ListDetail still reads all pages.
	Limit int
}

type createRequest struct {
	Flavor createBody `json:"flavor"`
}

type createBody struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	VCPUs            int    `json:"vcpus"`
	RAM              int    `json:"ram"`
	Disk             int    `json:"disk"`
	IsPublic         bool   `json:"os-flavor-access:is_public"`
	GPU              string `json:"gpu,omitempty"`
	DedicatedCPUs    *bool  `json:"dedicated_cpus,omitempty"`
	NetworkQoS       int    `json:"network_qos,omitempty"`
	AvailabilityZone string `json:"availability_zone,omitempty"`
	HighFreq         *bool  `json:"highfreq,omitempty"`
	UltraHighFreq    *bool  `json:"ultra_highfreq,omitempty"`
	VCPUsVendor      string `json:"vcpus_vendor,omitempty"`
}

type viewEnvelope struct {
	Flavor View `json:"flavor"`
}

type pageEnvelope struct {
	Flavors []View            `json:"flavors"`
	Links   []pagination.Link `json:"flavors_links"`
}

func (p *pageEnvelope) Items() []View { return p.Flavors }

func (p *pageEnvelope) NextHref() string { return pagination.NextHref(p.Links) }
