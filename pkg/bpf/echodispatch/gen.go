package echodispatch

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go echo_dispatch ./echo_dispatch.bpf.c -- -I/usr/local/include -I/usr/include
