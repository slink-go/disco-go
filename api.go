package disco_go

import "github.com/slink-go/disco/common/api"

type DiscoClient interface {

	// Leave - unregister client from disco server
	Leave() error

	// Registry - get registry object
	Registry() DiscoRegistry
}

type DiscoRegistry interface {

	// Services - list all registered services
	Services() []string

	// List - list all registered clients
	List() []Client

	// Get - get client for registered service
	//       (only in UP state; next from ring buffer)
	Get(service string) (Client, error)

	// Sync - synchronize registrations data
	Sync(clients []Client)

	// Reset - reset local discovery cache
	Reset()
}

type Client interface {
	ClientId() string
	ServiceId() string
	State() api.ClientState
	Endpoint(typ api.EndpointType) (string, error)
	Field(key string) (any, bool)
}

type serviceCall func(url string, data map[string]any, headers map[string]string) ([]byte, map[string]string, int, error)

const (
	joinUrlTemplate  = "%s/api/join"
	leaveUrlTemplate = "%s/api/leave?id=%s"
	pingUrlTemplate  = "%s/api/ping?id=%s"
	listUrlTemplate  = "%s/api/list"
)
