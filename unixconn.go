package memrpc

import (
	"github.com/justincormack/go-memfd"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"

	"errors"

	// Included for Sizeof *only*, for use in computing the size of our oob
	// buffer:
	"unsafe"
)

var (
	errShortRead      = errors.New("Short read")
	errWrongCmsgLen   = errors.New("Unexpected Cmsg len; should be 1")
	errWrongRightsLen = errors.New("Unexpected Rights len; should be 1")
)

type UnixMemfdConn struct {
	Fd int
}

func (conn *UnixMemfdConn) Send(ctx context.Context, memfd *memfd.Memfd) error {
	peername, err := unix.Getpeername(conn.Fd)
	if err != nil {
		return err
	}
	oob := unix.UnixRights(int(memfd.File.Fd()))
	return unix.Sendmsg(conn.Fd, []byte{0}, oob, peername, unix.MSG_OOB)
}

func (conn *UnixMemfdConn) Recv(ctx context.Context) (*memfd.Memfd, error) {
	p := []byte{0}
	oob := make([]byte, unix.CmsgSpace(int(unsafe.Sizeof(int(0)))))
	n, oobn, _, _, err := unix.Recvmsg(conn.Fd, p, oob, 0)
	if err != nil {
		return nil, err
	}
	if n != len(p) || oobn != len(oob) {
		return nil, errShortRead
	}

	ctlMsgs, err := unix.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, err
	}
	if len(ctlMsgs) != 1 {
		return nil, errWrongCmsgLen
	}
	rights, err := unix.ParseUnixRights(&ctlMsgs[0])
	if err != nil {
		return nil, err
	}
	if len(rights) != 1 {
		for _, fd := range rights {
			unix.Close(fd)
		}
		return nil, errWrongRightsLen
	}
	memfd, err := memfd.New(uintptr(rights[0]))
	if err != nil {
		unix.Close(rights[0])
		return nil, err
	}
	return memfd, nil
}

func (conn *UnixMemfdConn) Close() error {
	return unix.Close(conn.Fd)
}
