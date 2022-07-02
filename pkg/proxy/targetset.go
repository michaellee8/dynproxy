package proxy

import (
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"sync"
	"time"
)

const defaultRetryBackoffBase = 2
const defaultRetryBackoffFactor = 15 * time.Second
const defaultHealthcheckTimeout = 5 * time.Second
const defaultRetryMaxExponent = 8

type TargetSetWithHealthCheck struct {
	ts *targetset.TargetSet[string]

	retryBackoffBase   int
	retryBackoffFactor time.Duration
	maxExponent        int
	healthchecker      func(target string) bool

	closeWg *sync.WaitGroup
}

func NewTargetSetWithHealthCheck() *TargetSetWithHealthCheck {
	return NewTargetSetWithHealthCheckCustom(defaultRetryBackoffBase, defaultRetryBackoffFactor, defaultRetryMaxExponent, defaultHealthchecker)
}

func NewTargetSetWithHealthCheckCustom(
	backoffBase int,
	backoffFactor time.Duration,
	maxExponent int,
	healthchecker func(target string) bool,
) *TargetSetWithHealthCheck {
	return &TargetSetWithHealthCheck{
		ts:                 targetset.NewTargetSet[string](),
		retryBackoffBase:   backoffBase,
		retryBackoffFactor: backoffFactor,
		maxExponent:        maxExponent,
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
	hc.closeWg.Add(1)
	go hc.startChecker()
	return nil
}

func (hc *TargetSetWithHealthCheck) Close() (err error) {
	hc.closeWg.Done()
	return nil
}

func (hc *TargetSetWithHealthCheck) startChecker() {
	exit := make(chan struct{})
	go func() {
		hc.closeWg.Wait()
		exit <- struct{}{}
	}()

	for {
		select {
		case <-exit:
			return
		case <-time.Tick(hc.retryBackoffFactor):
			picker := hc.Picker()
			firstPick, err := picker.Pick()
			if err != nil {
				logrus.Error(errors.Wrap(err, "cannot pick target"))
				continue
			}
			hc.checkTarget(firstPick)
			for {
				pick, err := picker.Pick()
				if pick == firstPick {
					// break if we loop back to the first pick
					break
				}
				if err != nil {
					logrus.Error(errors.Wrap(err, "cannot pick target"))
					break
				}
				hc.checkTarget(pick)
			}
		}

	}
}

func (hc *TargetSetWithHealthCheck) checkTarget(target string) {
	if !hc.healthchecker(target) {
		hc.ts.Block(target)
		go hc.startHealthcheckRetry(target)
	}
}

func (hc *TargetSetWithHealthCheck) startHealthcheckRetry(target string) {

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
