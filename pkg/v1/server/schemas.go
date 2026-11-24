package server

import (
	"encoding/json"
	"time"

	"github.com/selectel/compute-go/pkg/v1/internal/pagination"
)

type View struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`

	// VMState and TaskState refine Status. TaskState is empty when no operation is in progress.
	VMState   string `json:"OS-EXT-STS:vm_state"`
	TaskState string `json:"OS-EXT-STS:task_state"`

	// Flavor is a copy of the flavor taken at boot or resize. It carries the flavor name and
	// no flavor ID.
	Flavor Flavor `json:"flavor"`

	// Image is empty for a server booted from a volume.
	Image Image `json:"image"`

	AvailabilityZone string `json:"OS-EXT-AZ:availability_zone"`
	KeyName          string `json:"key_name"`

	// ConfigDrive is reported by the API as an empty string or "True".
	ConfigDrive ConfigDrive `json:"config_drive"`

	// Addresses are keyed by the network name.
	Addresses  map[string][]Address `json:"addresses"`
	AccessIPv4 string               `json:"accessIPv4"`
	AccessIPv6 string               `json:"accessIPv6"`

	// Metadata does not include the service keys that the API hides.
	Metadata map[string]string `json:"metadata"`

	// Tags include the preemptible tag of a preemptible server.
	Tags []string `json:"tags"`

	// ServerGroups holds the UUID of the server group, if the server is a member of one.
	ServerGroups []string `json:"server_groups"`

	VolumesAttached []AttachedVolume `json:"os-extended-volumes:volumes_attached"`

	// PreemptibleForbiddenUntil is the moment until which the cloud refuses to make the server
	// preemptible again.
	PreemptibleForbiddenUntil ForbiddenUntil `json:"preemptible_forbidden_until"`

	// Fault is set when the server is in the ERROR status or an operation has failed.
	Fault *Fault `json:"fault"`

	// UserData is the Base64-encoded user data, empty for a server without one. Get
	// reports it; ListDetail does not.
	UserData string `json:"user_data"`

	UserID    string    `json:"user_id"`
	ProjectID string    `json:"tenant_id"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
}

// ForbiddenUntil is the end of the period in which the cloud refuses to make a server
// preemptible again. The cloud sets it asynchronously after the server stops being
// preemptible.
type ForbiddenUntil struct {
	// Reported is false when the pool sets no such period: the API then omits the field.
	// It is true when the API returns the field, even as null.
	Reported bool

	// Time is empty while no period applies to the server.
	Time string
}

// UnmarshalJSON records that the API returned the field, also as null.
func (until *ForbiddenUntil) UnmarshalJSON(raw []byte) error {
	until.Reported = true
	if string(raw) == "null" {
		until.Time = ""

		return nil
	}

	return json.Unmarshal(raw, &until.Time)
}

type Flavor struct {
	// OriginalName is the name of the flavor the server was booted or resized with.
	OriginalName string `json:"original_name"`
	VCPUs        int    `json:"vcpus"`

	// RAM is in MiB; Disk, Ephemeral and Swap follow the units of the flavor API.
	RAM        int               `json:"ram"`
	Disk       int               `json:"disk"`
	Ephemeral  int               `json:"ephemeral"`
	Swap       int               `json:"swap"`
	ExtraSpecs map[string]string `json:"extra_specs"`
}

type Image struct {
	ID string `json:"id"`
}

// UnmarshalJSON accepts the empty string that the API reports for a volume-backed server.
func (image *Image) UnmarshalJSON(raw []byte) error {
	if len(raw) != 0 && raw[0] == '"' {
		*image = Image{}

		return nil
	}

	type plain Image

	return json.Unmarshal(raw, (*plain)(image))
}

type ConfigDrive bool

func (drive *ConfigDrive) UnmarshalJSON(raw []byte) error {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}

	switch typed := value.(type) {
	case bool:
		*drive = ConfigDrive(typed)
	case string:
		*drive = typed == "True" || typed == "true"
	default:
		*drive = false
	}

	return nil
}

type Address struct {
	Address string `json:"addr"`
	Version int    `json:"version"`

	// Type is fixed or floating.
	Type string `json:"OS-EXT-IPS:type"`
	MAC  string `json:"OS-EXT-IPS-MAC:mac_addr"`
}

type AttachedVolume struct {
	ID                  string `json:"id"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

type Fault struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Created is the answer to a create request. The server is not built yet.
type Created struct {
	ID string `json:"id"`
}

// Interface is a port attached to the server.
type Interface struct {
	PortID    string    `json:"port_id"`
	NetworkID string    `json:"net_id"`
	MAC       string    `json:"mac_addr"`
	PortState string    `json:"port_state"`
	FixedIPs  []FixedIP `json:"fixed_ips"`
}

type FixedIP struct {
	Address  string `json:"ip_address"`
	SubnetID string `json:"subnet_id"`
}

type viewEnvelope struct {
	Server View `json:"server"`
}

type createdEnvelope struct {
	Server Created `json:"server"`
}

type metadataEnvelope struct {
	Metadata map[string]string `json:"metadata"`
}

type tagsEnvelope struct {
	Tags []string `json:"tags"`
}

type interfacesEnvelope struct {
	Interfaces []Interface `json:"interfaceAttachments"`
}

type pageEnvelope struct {
	Servers []View            `json:"servers"`
	Links   []pagination.Link `json:"servers_links"`
}

func (p *pageEnvelope) Items() []View { return p.Servers }

func (p *pageEnvelope) NextHref() string { return pagination.NextHref(p.Links) }
