package client

import (
	"container/ring"
	"fmt"
	disco "github.com/slink-go/disco/common/api"
	"github.com/slink-go/logger"
	"strings"
	"sync"
)

func NewRegistry() DiscoRegistry {
	return &registry{}
}

type registry struct {
	sync.RWMutex
	clientList []Client
	clientRing map[string]*ring.Ring
}

func (r *registry) Services() []string {
	r.RLock()
	defer r.RUnlock()
	m := map[string]struct{}{}
	for _, v := range r.clientList {
		m[v.ServiceId()] = struct{}{}
	}
	var result []string
	for k := range m {
		result = append(result, k)
	}
	return result
}
func (r *registry) List() []Client {
	r.RLock()
	defer r.RUnlock()
	return r.clientList
}
func (r *registry) Get(service string) (Client, error) {
	service = strings.ToUpper(service)
	r.RLock()
	defer r.RUnlock()
	v, ok := r.clientRing[service]
	if !ok {
		return nil, fmt.Errorf("no clients found for %s", service)
	}
	r.clientRing[service] = r.clientRing[service].Next()
	return v.Value.(Client), nil
}
func (r *registry) Sync(clients []Client) {

	r.Lock()
	defer r.Unlock()

	for _, v := range clients {
		logger.Debug("[sync] %s %s", v.ClientId(), v.State())
	}

	r.clientList = clients

	svcs := map[string][]Client{}
	for i := 0; i < len(clients); i++ {
		if clients[i].State() == disco.ClientStateUp {
			svcs[clients[i].ServiceId()] = append(svcs[clients[i].ServiceId()], clients[i])
		}
	}

	// fill in clients ring (used for client-side load balancing)
	r.clientRing = make(map[string]*ring.Ring)
	for k, v := range svcs {
		r.clientRing[k] = ring.New(len(v))
		for _, c := range v {
			r.clientRing[k].Value = c
			r.clientRing[k] = r.clientRing[k].Next()
		}
	}

}
