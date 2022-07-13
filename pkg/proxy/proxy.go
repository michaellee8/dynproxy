package proxy

import (
	"fmt"
	"github.com/michaellee8/dynproxy/pkg/bpf/echodispatch"
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"github.com/michaellee8/dynproxy/pkg/proxy/op"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

type DynProxy struct {
	// opMut is the mutex to be held when applying operations
	opMut *sync.RWMutex

	portMap map[int]*portMapValue

	ruleMap map[string]*ruleMapValue

	ruleTargetMap map[ruleTargetMapKey]*ruleTargetMapValue

	echoDispatch *echodispatch.EchoDispatch

	ebpf bool

	logger *logrus.Entry

	// used if ebpf support is enabled
	lis net.Listener
}

func NewDynProxy(ebpf bool) *DynProxy {
	return &DynProxy{
		opMut:         &sync.RWMutex{},
		portMap:       make(map[int]*portMapValue),
		ruleMap:       make(map[string]*ruleMapValue),
		ruleTargetMap: make(map[ruleTargetMapKey]*ruleTargetMapValue),
		echoDispatch:  echodispatch.NewEchoDispatch(),
		ebpf:          ebpf,
		logger:        logrus.WithField("src", "dynproxy"),
	}
}

func (p *DynProxy) hasRule(rule string) bool {
	_, ok := p.ruleMap[rule]
	return ok
}

func (p *DynProxy) HasRule(rule string) bool {
	p.opMut.RLock()
	defer p.opMut.RUnlock()
	return p.hasRule(rule)
}

func (p *DynProxy) applyOperation(op op.Operation) (err error) {

}

func (p *DynProxy) ApplyOperation(op op.Operation) (err error) {
	p.opMut.Lock()
	defer p.opMut.Unlock()
	return p.applyOperation(op)
}

func (p *DynProxy) addTarget(rule string, target string) (err error) {
	if !p.hasRule(rule) {
		return errors.Wrap(ErrRuleNotExist, "unable to add target")
	}
	rv := p.ruleMap[rule]
	if rv.targetSet.Has(target) {
		return errors.Wrap(ErrTargetAlreadyExist, "unable to add target")
	}
	p.logger.Infof("adding target %s to rule %s", target, rule)
	rv.targetSet.Add(target)
	p.ruleTargetMap[ruleTargetMapKey{rule: rule, target: target}] = &ruleTargetMapValue{
		mut:     &sync.RWMutex{},
		connSet: &gsync.MapOf[*net.TCPConn, struct{}]{},
	}
	return nil
}

func (p *DynProxy) removeTarget(rule string, target string) (err error) {
	if !p.hasRule(rule) {
		return errors.Wrap(ErrRuleNotExist, "unable to remove target")
	}
}

var ErrRuleNotExist = errors.New("rule does not exist")
var ErrRuleAlreadyExist = errors.New("rule already exists")
var ErrInternalIntegrity = errors.New("fatal error: DynProxy internal integrity failure")
var ErrTargetNotExist = errors.New("target does not exist for the rule")
var ErrPortNotExist = errors.New("port does not exist")
var ErrTargetAlreadyExist = errors.New("target already exist")
var ErrPortAlreadyExist = errors.New("port already exist")

type VerificationError struct {
	fieldName string
}

func (e *VerificationError) Error() string {
	return fmt.Sprintf("invalid field %s", e.fieldName)
}

func errIsUnexpected(err error) bool {
	return !(err == nil || errors.Is(err, net.ErrClosed))
}
