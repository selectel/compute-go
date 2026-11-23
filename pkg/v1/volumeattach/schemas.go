package volumeattach

type View struct {
	ServerID string `json:"serverId"`
	VolumeID string `json:"volumeId"`

	// Device is the device name chosen by the cloud, such as /dev/sdb.
	Device string `json:"device"`

	// Tag is the device tag that the server sees in its device metadata.
	Tag                 string `json:"tag"`
	DeleteOnTermination bool   `json:"delete_on_termination"`
}

type CreateOpts struct {
	VolumeID string
	Tag      string
}

type createRequest struct {
	VolumeAttachment createBody `json:"volumeAttachment"`
}

type createBody struct {
	VolumeID string `json:"volumeId"`
	Tag      string `json:"tag,omitempty"`
}

type viewEnvelope struct {
	VolumeAttachment View `json:"volumeAttachment"`
}

type listEnvelope struct {
	VolumeAttachments []View `json:"volumeAttachments"`
}
