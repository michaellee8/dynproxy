//go:build ebpf
// +build ebpf

package echodispatch

import (
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/rlimit"
	"github.com/pkg/errors"
	"os"
)

type EchoDispatch struct {
	objs  echo_dispatchObjects
	netns *os.File
	link  *link.NetNsLink
}

func NewEchoDispatch() *EchoDispatch {
	return &EchoDispatch{}
}

func (ed *EchoDispatch) Load() (err error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return errors.Wrap(err, "cannot load ebpf program, cannot remove memlock")
	}

	ed.objs = echo_dispatchObjects{}
	if err := loadEcho_dispatchObjects(&ed.objs, nil); err != nil {
		return errors.Wrap(err, "cannot load ebpf object")
	}

	if ed.netns, err = os.Open("/proc/self/ns/net"); err != nil {
		return errors.Wrap(err, "cannot get ns")
	}

	if ed.link, err = link.AttachNetNs(int(ed.netns.Fd()), ed.objs.EchoDispatch); err != nil {
		return errors.Wrap(err, "cannot attach program")
	}
	return nil
}

func (ed *EchoDispatch) SetSocketFd(fd uintptr) (err error) {
	if err := ed.objs.EchoSocket.Put(uint32(0), uint64(fd)); err != nil {
		return errors.Wrap(err, "cannot set socket fd to ebpf map")
	}
	return nil
}

func (ed *EchoDispatch) AddPort(port int) (err error) {
	if err := ed.objs.EchoPorts.Put(uint16(port), uint8(0)); err != nil {
		return errors.Wrap(err, "cannot add port to echo ports")
	}
	return nil
}

func (ed *EchoDispatch) RemovePort(port int) (err error) {
	if err := ed.objs.EchoPorts.Delete(uint16(port)); err != nil {
		return errors.Wrap(err, "cannot delete port from echo ports")
	}
	return nil
}

func (ed *EchoDispatch) Close() (err error) {
	if err := ed.link.Close(); err != nil {
		return errors.Wrap(err, "cannot close link")
	}
	if err := ed.netns.Close(); err != nil {
		return errors.Wrap(err, "cannot close ns file")
	}
	if err := ed.objs.Close(); err != nil {
		return errors.Wrap(err, "cannot close ebpf objs")
	}
	return nil
}

func (ed *EchoDispatch) Supported() bool {
	return true
}
