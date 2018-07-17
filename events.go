package main

import (
	"net"
	"reflect"
	"time"
)

// ToxPollEvent
// ToxReadyReadEvent
// KcpPollEvent
// KcpReadyReadEvent
// ClientReadyReadEvent
// NewConnEvent

type ToxPollEvent struct{}
type ToxReadyReadEvent struct {
	friendNumber uint32
	message      string
}
type ToxMessageEvent ToxReadyReadEvent

type KcpPollEvent struct{}
type KcpReadyReadEvent struct {
	ch *Channel
}
type KcpOutputEvent struct {
	buf   []byte
	size  int
	extra interface{}
}
type KcpCheckCloseEvent struct{}

type NewConnEvent struct {
	conn  net.Conn
	times int
	btime time.Time
	tname string
}

// TODO tox prefix?
type ClientReadyReadEvent struct {
	ch   *Channel
	buf  []byte
	size int
	ctrl bool
}
type ServerReadyReadEvent ClientReadyReadEvent

type ClientCloseEvent struct {
	ch *Channel
}
type ServerCloseEvent ClientCloseEvent
type ClientCheckACKEvent struct {
	ch *Channel
}
type ChannelGCEvent struct{}

type UdpReadyReadEvent struct {
	addr net.Addr
	buf  []byte
	size int
}

type CommonEvent struct {
	t reflect.Type
	v reflect.Value
}
