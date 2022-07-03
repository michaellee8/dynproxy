package proxy

import (
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"net"
	"sync"
)
import gsync "github.com/SaveTheRbtz/generic-sync-map-go"

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
