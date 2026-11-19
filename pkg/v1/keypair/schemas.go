package keypair

import "github.com/selectel/compute-go/pkg/v1/internal/pagination"

// View describes a key pair of the user that owns the token.
type View struct {
	Name        string `json:"name"`
	PublicKey   string `json:"public_key"`
	Fingerprint string `json:"fingerprint"`

	// Type is ssh or x509.
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

type createRequest struct {
	KeyPair createBody `json:"keypair"`
}

type createBody struct {
	Name      string `json:"name"`
	PublicKey string `json:"public_key"`
}

type viewEnvelope struct {
	KeyPair View `json:"keypair"`
}

type pageEnvelope struct {
	KeyPairs []viewEnvelope    `json:"keypairs"`
	Links    []pagination.Link `json:"keypairs_links"`
}

func (p *pageEnvelope) Items() []View {
	items := make([]View, 0, len(p.KeyPairs))
	for _, item := range p.KeyPairs {
		items = append(items, item.KeyPair)
	}

	return items
}

func (p *pageEnvelope) NextHref() string { return pagination.NextHref(p.Links) }
