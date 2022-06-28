package tcpecho

import (
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	"testing"
)

func TestTCPEchoer_NoShout(t *testing.T) {
	testEchoer(t, false)
}

func TestTCPEchoer_Shout(t *testing.T) {
	testEchoer(t, true)
}

func testEchoer(t *testing.T, shout bool) {
	te := NewTCPEchoer(0, shout)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-te.errChan:
				if errors.Is(err, net.ErrClosed) {
					continue
				}
				t.Log(err)
			case <-done:
				return
			}
		}
	}()
	defer func() {
		done <- struct{}{}
	}()
	require.NoError(t, te.Start())
	defer func() {
		require.NoError(t, te.Close())
	}()
	port, err := te.Port()
	require.NoError(t, err)
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{Port: port})
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, conn.Close())
	}()
	go func() {
		_, err := conn.Write([]byte("aaaaabbbbb"))
		require.NoError(t, err)
	}()
	lr := io.LimitReader(conn, 10)
	rbs, err := io.ReadAll(lr)
	require.NoError(t, err)
	if shout {
		require.Equal(t, "AAAAABBBBB", string(rbs))
	} else {
		require.Equal(t, "aaaaabbbbb", string(rbs))
	}

}
