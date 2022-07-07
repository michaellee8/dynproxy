package proxy

import (
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"net"
	"sync"
)
import gsync "github.com/SaveTheRbtz/generic-sync-map-go"

// Good reference for connection close handling:
// https://github.com/mholt/caddy-l4/blob/56bd7700d889f2ffd52353c241819ddbc7745ff6/modules/l4proxy/proxy.go#L281

type ruleTargetMapKey struct {
	rule   string
	target string
}

type portMapValue struct {
	rule    string
	lis     *net.Listener
	connSet *gsync.MapOf[*net.TCPConn, struct{}]
	mut     *sync.RWMutex
}

type ruleMapValue struct {
	targetSet *TargetSetWithHealthCheck
	picker    *targetset.Picker[string]
}

type ruleTargetMapValue struct {
	connSet *gsync.MapOf[*net.TCPConn, struct{}]
	mut     *sync.RWMutex
}
