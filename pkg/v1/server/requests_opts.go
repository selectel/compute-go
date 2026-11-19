package server

import "encoding/json"

// Network modes that replace the list of networks.
const (
	NetworkModeAuto = "auto"
	NetworkModeNone = "none"
)

type CreateOpts struct {
	Name string

	// FlavorID is required. ImageID is empty for a server booted from a volume.
	FlavorID string
	ImageID  string

	AvailabilityZone string
	KeyName          string
	ConfigDrive      bool

	// UserData is sent as is and must be Base64-encoded.
	UserData string

	// AdminPass is applied by images that support it, such as Windows images.
	AdminPass string
	Metadata  map[string]string

	// Tags make the server preemptible when they include the preemptible tag.
	Tags []string

	// Networks and NetworkMode are mutually exclusive, and one of them is required.
	Networks    []Network
	NetworkMode string

	BlockDevices []BlockDevice

	// ServerGroupID places the server in a server group.
	ServerGroupID string
}

// Network needs a network UUID or a port UUID.
type Network struct {
	UUID    string `json:"uuid,omitempty"`
	Port    string `json:"port,omitempty"`
	FixedIP string `json:"fixed_ip,omitempty"`
}

type BlockDevice struct {
	// SourceType is blank, image, snapshot or volume; DestinationType is local or volume.
	SourceType      string `json:"source_type"`
	DestinationType string `json:"destination_type,omitempty"`

	// UUID identifies the source and is empty for a blank device.
	UUID string `json:"uuid,omitempty"`

	// BootIndex is 0 for the boot device and negative for a device that is not bootable.
	BootIndex int `json:"boot_index"`

	// VolumeSize is in GiB.
	VolumeSize          int    `json:"volume_size,omitempty"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
	GuestFormat         string `json:"guest_format,omitempty"`
	VolumeType          string `json:"volume_type,omitempty"`
	DeviceType          string `json:"device_type,omitempty"`
	DiskBus             string `json:"disk_bus,omitempty"`
}

// UpdateOpts change the server in place. Nil fields are not changed.
type UpdateOpts struct {
	Name *string

	// UserData replaces the user data with a Base64-encoded value. ClearUserData removes it.
	UserData      *string
	ClearUserData bool
}

func (opts UpdateOpts) isEmpty() bool {
	return opts.Name == nil && opts.UserData == nil && !opts.ClearUserData
}

// ListOpts are server-side filters. Empty fields are not sent.
type ListOpts struct {
	// Name is a regular expression matched against the server name.
	Name   string
	Status string

	// Limit is the page size. ListDetail still reads all pages.
	Limit int
}

type RebuildOpts struct {
	ImageID string
}

type createRequest struct {
	Server         createBody      `json:"server"`
	SchedulerHints *schedulerHints `json:"os:scheduler_hints,omitempty"`
}

type schedulerHints struct {
	Group string `json:"group"`
}

type createBody struct {
	Name             string            `json:"name"`
	FlavorRef        string            `json:"flavorRef"`
	ImageRef         string            `json:"imageRef,omitempty"`
	AvailabilityZone string            `json:"availability_zone,omitempty"`
	KeyName          string            `json:"key_name,omitempty"`
	ConfigDrive      bool              `json:"config_drive,omitempty"`
	UserData         string            `json:"user_data,omitempty"`
	AdminPass        string            `json:"adminPass,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Tags             []string          `json:"tags,omitempty"`

	// Networks is the mode string or the list of networks.
	Networks     any           `json:"networks"`
	BlockDevices []BlockDevice `json:"block_device_mapping_v2,omitempty"`
}

type updateRequest struct {
	Server map[string]any `json:"server"`
}

// action encodes {"<name>": <body>}; a nil body is sent as null.
type action map[string]any

func (a action) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(a))
}

type resizeBody struct {
	FlavorRef string `json:"flavorRef"`
}

type rebuildBody struct {
	ImageRef string `json:"imageRef"`
}

type changePasswordBody struct {
	AdminPass string `json:"adminPass"`
}
