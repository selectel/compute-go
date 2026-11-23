package servergroup

type View struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Policy is anti-affinity, soft-anti-affinity, affinity or soft-affinity.
	Policy string `json:"policy"`
	Rules  Rules  `json:"rules"`

	// Members are the UUIDs of the servers in the group.
	Members   []string `json:"members"`
	ProjectID string   `json:"project_id"`
	UserID    string   `json:"user_id"`
}

type Rules struct {
	// MaxServerPerHost is 0 when the rule is not set.
	MaxServerPerHost int `json:"max_server_per_host,omitempty"`
}

type CreateOpts struct {
	Name   string
	Policy string

	// MaxServerPerHost is sent only when it is positive; the API accepts it for anti-affinity.
	MaxServerPerHost int
}

type createRequest struct {
	ServerGroup createBody `json:"server_group"`
}

type createBody struct {
	Name   string `json:"name"`
	Policy string `json:"policy"`
	Rules  *Rules `json:"rules,omitempty"`
}

type viewEnvelope struct {
	ServerGroup View `json:"server_group"`
}

type listEnvelope struct {
	ServerGroups []View `json:"server_groups"`
}
