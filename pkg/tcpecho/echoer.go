package tcpecho

import (
	"github.com/pkg/errors"
	"io"
	"net"
)

// TCPEchoer providers a TCP server that echoes whatever data sent to it. It transforms the
// english letter in the data to uppercase if the `shout` option is enabled.
type TCPEchoer struct {
	port  int
	shout bool

	errChan chan error

	lis *net.TCPListener
}

func NewTCPEchoer(port int, shout bool) *TCPEchoer {
	return &TCPEchoer{
		port:    port,
		shout:   shout,
		errChan: make(chan error, 100),
	}
}

func (e *TCPEchoer) Start() (err error) {
	if e.lis != nil {
		return errors.Wrap(ErrAlreadyStarted, "cannot start")
	}
	e.lis, err = net.ListenTCP("tcp", &net.TCPAddr{Port: e.port})
	if err != nil {
		return errors.Wrap(err, "cannot listen tcp")
	}
	for {
		conn, err := e.lis.AcceptTCP()
		if err != nil {
			return errors.Wrap(err, "cannot listen connection")
		}
		go e.handleConn(conn)
	}
}

func (e *TCPEchoer) Port() (port int, err error) {
	if e.lis == nil {
		return 0, errors.Wrap(ErrNotStarted, "cannot get port")
	}
	addr, ok := e.lis.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.Wrap(errAssertion, "cannot get port")
	}
	return addr.Port, nil
}

func (e *TCPEchoer) Close() (err error) {
	if err := e.lis.Close(); err != nil {
		return errors.Wrap(err, "cannot close TCPEchoer")
	}
	return nil
}

func (e *TCPEchoer) ErrChan() <-chan error {
	return e.errChan
}

func (e *TCPEchoer) handleConn(conn *net.TCPConn) {
	if _, err := e.copy(conn, conn); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			e.errChan <- err
		}
		_ = conn.Close()
		return
	}
}

func (e *TCPEchoer) copy(dst io.Writer, src io.Reader) (written int64, err error) {
	if e.shout {
		return io.Copy(dst, src)
	} else {
		return copyShout(dst, src)
	}
}

// copyShout copy the bytes read from src to dst, and transform all lowercase letter from
// src to uppercase letter, assuming that the whole byte stream are ascii encoded.
func copyShout(dst io.Writer, src io.Reader) (written int64, err error) {
	const size = 32 * 1024

	const capLetterDiff = byte('A') - byte('a')

	buf := make([]byte, size)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			for i := 0; i < nr; i++ {
				if 'a' <= buf[i] && buf[i] <= 'z' {
					buf[i] += capLetterDiff
				}
			}
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

var errInvalidWrite = errors.New("invalid write result")
var ErrNotStarted = errors.New("not started")
var ErrAlreadyStarted = errors.New("already started")
var errAssertion = errors.New("assertion failed")
