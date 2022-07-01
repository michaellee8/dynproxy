//go:build !ebpf

package echodispatch

import (
	"github.com/pkg/errors"
)

type EchoDispatch struct {
}

func NewEchoDispatch() *EchoDispatch {
	return &EchoDispatch{}
}

func (ed *EchoDispatch) Load() (err error) {
	return errNotSupported
}

func (ed *EchoDispatch) SetSocketFd(fd uintptr) (err error) {
	return errNotSupported
}

func (ed *EchoDispatch) AddPort(port int) (err error) {
	return errNotSupported
}

func (ed *EchoDispatch) RemovePort(port int) (err error) {
	return errNotSupported
}

func (ed *EchoDispatch) Close() (err error) {
	return errNotSupported
}

func (ed *EchoDispatch) Supported() bool {
	return false
}

var errNotSupported = errors.New("not supported")
