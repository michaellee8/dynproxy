package proxy

import (
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"net"
	"sync"
	"time"
)

const defaultRetryBackoffBase = 2
const defaultRetryBackoffFactor = 5 * time.Second
const defaultHealthcheckTimeout = 15 * time.Second

type TargetSetWithHealthCheck struct {
	ts *targetset.TargetSet[string]

	retryBackoffBase   int
	retryBackoffFactor time.Duration
	healthchecker      func(target string) bool

	closeWg *sync.WaitGroup
}

func NewTargetSetWithHealthCheck() *TargetSetWithHealthCheck {
	return NewTargetSetWithHealthCheckCustom(defaultRetryBackoffBase, defaultRetryBackoffFactor, defaultHealthchecker)
}

func NewTargetSetWithHealthCheckCustom(backoffBase int, backoffFactor time.Duration, healthchecker func(target string) bool) *TargetSetWithHealthCheck {
	return &TargetSetWithHealthCheck{
		ts:                 targetset.NewTargetSet[string](),
		retryBackoffBase:   backoffBase,
		retryBackoffFactor: backoffFactor,
		healthchecker:      healthchecker,
		closeWg:            &sync.WaitGroup{},
	}
}

func (hc *TargetSetWithHealthCheck) Len() int {
	return hc.ts.Len()
}

func (hc *TargetSetWithHealthCheck) Picker() *targetset.Picker[string] {
	return targetset.NewPicker(hc.ts)
}

func (hc *TargetSetWithHealthCheck) AllPicker() *targetset.Picker[string] {
	return targetset.NewAllPicker(hc.ts)
}

func (hc *TargetSetWithHealthCheck) Add(target string) bool {
	return hc.ts.Add(target)
}

func (hc *TargetSetWithHealthCheck) Remove(target string) bool {
	return hc.ts.Remove(target)
}

func (hc *TargetSetWithHealthCheck) Has(target string) bool {
	return hc.ts.Has(target)
}

func (hc *TargetSetWithHealthCheck) HasUnblocked(target string) bool {
	return hc.ts.HasUnblocked(target)
}

func (hc *TargetSetWithHealthCheck) IsBlocked(target string) bool {
	return hc.ts.IsBlocked(target)
}

func (hc *TargetSetWithHealthCheck) Start() (err error) {

}

func (hc *TargetSetWithHealthCheck) Close() (err error) {
	hc.closeWg.Done()
}

func defaultHealthchecker(target string) bool {
	conn, err := net.DialTimeout("tcp", target, defaultHealthcheckTimeout)
	if err != nil {
		return false
	}
	defer func() {
		_ = conn.Close()
	}()
	return true
}
