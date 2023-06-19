package client

import (
	"github.com/ws-slink/disco/common/api"
	"strings"
)

// region - registered client

type client struct {
	//Tenant    string          `json:"tenant,omitempty"`
	ClientId_  string          `json:"client_id"`
	ServiceId_ string          `json:"service_id"`
	Endpoints_ []endpoint      `json:"endpoints,omitempty"`
	Meta_      map[string]any  `json:"meta,omitempty"`
	State_     api.ClientState `json:"state"`
}

func (c *client) ClientId() string {
	return c.ClientId_
}
func (c *client) ServiceId() string {
	return c.ServiceId_
}
func (c *client) State() api.ClientState {
	return c.State_
}
func (c *client) Endpoint(typ api.EndpointType) (string, error) {
	for _, v := range c.Endpoints_ {
		if v.Type() == typ {
			return v.Url, nil
		}
	}
	return "", nil
}
func (c *client) Field(key string) (v any, ok bool) {
	v, ok = c.Meta_[key]
	return
}

// endregion
// region - endpoint

type endpoint struct {
	Url string `json:"url"`
}

func (e *endpoint) Type() api.EndpointType {
	s := strings.ToLower(e.Url)
	switch {
	case strings.HasPrefix(s, "https://"):
		return api.HttpsEndpoint
	case strings.HasPrefix(s, "http://"):
		return api.HttpEndpoint
	case strings.HasPrefix(s, "grpc://"):
		return api.GrpcEndpoint
	}
	return api.UnknownEndpoint
}

// endregion
