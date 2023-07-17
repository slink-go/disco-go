package disco_go

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

type ringBuffer struct {
	sync.RWMutex
	clientRing map[string]*ring.Ring
}

func newClientRing() *ringBuffer {
	return &ringBuffer{
		clientRing: make(map[string]*ring.Ring),
	}
}
func (b *ringBuffer) Get(serviceId string) (*ring.Ring, bool) {
	b.RLock()
	v, ok := b.clientRing[serviceId]
	b.RUnlock()
	return v, ok
}
func (b *ringBuffer) New(serviceId string, size int) {
	b.Lock()
	b.clientRing[serviceId] = ring.New(size)
	b.Unlock()
}
func (b *ringBuffer) Set(serviceId string, client Client) {
	b.Lock()
	b.clientRing[serviceId].Value = client
	b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	b.Unlock()
}
func (b *ringBuffer) Next(serviceId string) (*ring.Ring, bool) {
	b.Lock()
	v, ok := b.clientRing[serviceId]
	b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	b.Unlock()
	return v, ok
}

type registry struct {
	sync.RWMutex
	clientList []Client
	clientRing *ringBuffer
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
	v, ok := r.clientRing.Get(service)
	if !ok {
		return nil, fmt.Errorf("no clients found for %s", service)
	}
	r.clientRing.Next(service)
	return v.Value.(Client), nil
}
func (r *registry) Sync(clients []Client) {

	r.Lock()
	defer r.Unlock()

	for _, v := range clients {
		logger.Debug("[sync] %s (%s) %s", v.ClientId(), v.ServiceId(), v.State())
	}

	r.clientList = clients

	svcs := map[string][]Client{}
	for i := 0; i < len(clients); i++ {
		if clients[i].State() == disco.ClientStateUp {
			svcs[clients[i].ServiceId()] = append(svcs[clients[i].ServiceId()], clients[i])
		}
	}

	// fill in clients ring (used for client-side load balancing)
	r.clientRing = newClientRing()
	for k, v := range svcs {
		r.clientRing.New(k, len(v))
		for _, c := range v {
			r.clientRing.Set(k, c)
		}
	}

}
func (r *registry) Reset() {
	r.clientList = []Client{}
	r.clientRing = newClientRing()
}
