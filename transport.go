package memrpc

import (
	"errors"
	"github.com/justincormack/go-memfd"
	"github.com/justincormack/go-memfd/memproto"
	"golang.org/x/net/context"
	"zombiezen.com/go/capnproto2"
	"zombiezen.com/go/capnproto2/std/capnp/rpc"
)

var (
	errMessageNotMemfdBacked = errors.New("Message is not backed by a Memfd.")
)

// A MemfdConn allows sending and receiving Memfds.
type MemfdConn interface {
	Send(ctx context.Context, memfd *memfd.Memfd) error
	Recv(ctx context.Context) (*memfd.Memfd, error)
	Close() error
}

// A capnproto rpc transport using Memfds. All messages must be backed by Memfds.
type MemfdTransport struct {
	// Underlying mechanism used to transfer Memfds
	Conn MemfdConn
}

func (t *MemfdTransport) SendMessage(ctx context.Context, msg rpc.Message) error {
	arena, ok := msg.Struct.Segment().Message().Arena.(*memproto.MemfdArena)
	if !ok {
		return errMessageNotMemfdBacked
	}
	return t.Conn.Send(ctx, arena.Memfd)
}

func (t *MemfdTransport) RecvMessage(ctx context.Context) (rpc.Message, error) {
	memfd, err := t.Conn.Recv(ctx)
	if err != nil {
		return rpc.Message{}, err
	}
	return rpc.ReadRootMessage(&capnp.Message{
		Arena: &memproto.MemfdArena{memfd},
	})
}

func (t *MemfdTransport) Close() error {
	return t.Conn.Close()
}
