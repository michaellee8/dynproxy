package proxy

import (
	"github.com/michaellee8/dynproxy/pkg/ds/targetset"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const defaultRetryBackoffFactor = 15 * time.Second
const defaultHealthcheckTimeout = 5 * time.Second
const defaultRetryMaxExponent = 8

// nice reference for cancellation: https://go.dev/blog/pipelines

type TargetSetWithHealthCheck struct {
	ts *targetset.TargetSet[string]

	retryBackoffFactor time.Duration
	maxExponent        int
	healthchecker      func(target string) bool

	closeChan chan struct{}
}

func NewTargetSetWithHealthCheck() *TargetSetWithHealthCheck {
	return NewTargetSetWithHealthCheckCustom(defaultRetryBackoffFactor, defaultRetryMaxExponent, defaultHealthchecker)
}

func NewTargetSetWithHealthCheckCustom(
	backoffFactor time.Duration,
	maxExponent int,
	healthchecker func(target string) bool,
) *TargetSetWithHealthCheck {
	return &TargetSetWithHealthCheck{
		ts:                 targetset.NewTargetSet[string](),
		retryBackoffFactor: backoffFactor,
		maxExponent:        maxExponent,
		healthchecker:      healthchecker,
		closeChan:          make(chan struct{}),
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

func (hc *TargetSetWithHealthCheck) PickerNoRepeat() *targetset.Picker[string] {
	return targetset.NewPickerNoRepeat(hc.ts)
}

func (hc *TargetSetWithHealthCheck) AllPickerNoRepeat() *targetset.Picker[string] {
	return targetset.NewAllPickerNoRepeat(hc.ts)
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
	go hc.startChecker()
	return nil
}

func (hc *TargetSetWithHealthCheck) Close() (err error) {
	close(hc.closeChan)
	return nil
}

func (hc *TargetSetWithHealthCheck) Block(target string) (err error) {
	hc.ts.Block(target)
	go hc.startHealthcheckRetry(target)
	return nil
}

func (hc *TargetSetWithHealthCheck) startChecker() {

	for {
		select {
		case <-hc.closeChan:
			return
		case <-time.Tick(hc.retryBackoffFactor):
			picker := hc.PickerNoRepeat()
			for {
				picked, err := picker.Pick()
				if err != nil {
					if !errors.Is(err, targetset.ErrArrivedEnd) {
						logrus.Error(errors.Wrap(err, "cannot pick target"))
					}
					break
				}
				go hc.checkTarget(picked)
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
	exponent := 0

	for {
		select {
		case <-time.After(hc.retryBackoffFactor * (1 << exponent)):
			if !hc.ts.IsBlocked(target) || !hc.ts.Has(target) {
				// stop the coroutine if target is no longer blocked
				return
			}

			if hc.healthchecker(target) {
				hc.ts.Unblock(target)
				return
			}

			if exponent < hc.maxExponent {
				exponent++
			}
		case <-hc.closeChan:
			return
		}
	}
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
