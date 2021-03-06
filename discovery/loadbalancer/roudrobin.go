package loadbalancer

import (
	"sync/atomic"

	"github.com/mouadino/go-nano/discovery"
)

type roundRobinLoadBalancer struct {
	mod uint64
}

// NewRoundRobin returns a loadbalancer strategy that choose an endpoint using
// round-robin algorithm.
func NewRoundRobin() *roundRobinLoadBalancer {
	return &roundRobinLoadBalancer{0}
}

func (s *roundRobinLoadBalancer) Endpoint(instances []discovery.Instance) (string, error) {
	if len(instances) == 0 {
		return "", NoEndpointError
	}
	instance := instances[s.mod%uint64(len(instances))]
	var old uint64
	for {
		old = atomic.LoadUint64(&s.mod)
		if atomic.CompareAndSwapUint64(&s.mod, old, old+1) {
			break
		}
	}
	return instance.Endpoint, nil
}
