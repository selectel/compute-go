package availabilityzone

type View struct {
	Name  string `json:"zoneName"`
	State State  `json:"zoneState"`
}

type State struct {
	Available bool `json:"available"`
}

type listEnvelope struct {
	Zones []View `json:"availabilityZoneInfo"`
}
