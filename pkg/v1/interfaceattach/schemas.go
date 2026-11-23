package interfaceattach

type View struct {
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

// CreateOpts need exactly one of PortID and NetworkID. FixedIP is allowed only with NetworkID.
type CreateOpts struct {
	PortID    string
	NetworkID string
	FixedIP   string
}

type createRequest struct {
	InterfaceAttachment createBody `json:"interfaceAttachment"`
}

type createBody struct {
	PortID    string           `json:"port_id,omitempty"`
	NetworkID string           `json:"net_id,omitempty"`
	FixedIPs  []requestFixedIP `json:"fixed_ips,omitempty"`
}

type requestFixedIP struct {
	Address string `json:"ip_address"`
}

type viewEnvelope struct {
	InterfaceAttachment View `json:"interfaceAttachment"`
}
