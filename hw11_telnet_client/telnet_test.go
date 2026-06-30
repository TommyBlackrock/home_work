package main

import (
	"bytes"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})
}

func TestTelnetClientNotConnected(t *testing.T) {
	client := NewTelnetClient(
		"127.0.0.1:1",
		time.Second,
		io.NopCloser(bytes.NewReader(nil)),
		io.Discard,
	)

	require.ErrorIs(t, client.Send(), ErrNotConnected)
	require.ErrorIs(t, client.Receive(), ErrNotConnected)
	require.NoError(t, client.Close())
	require.NoError(t, client.Close())
}

func TestTelnetClientCloseUnblocksSend(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer func() { require.NoError(t, listener.Close()) }()

	serverDone := make(chan error, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			serverDone <- acceptErr
			return
		}

		serverDone <- conn.Close()
	}()

	inputReader, inputWriter := io.Pipe()
	defer func() { require.NoError(t, inputWriter.Close()) }()

	client := NewTelnetClient(listener.Addr().String(), time.Second, inputReader, io.Discard)
	require.NoError(t, client.Connect())

	sendDone := make(chan error, 1)
	go func() {
		sendDone <- client.Send()
	}()

	require.NoError(t, client.Receive())
	require.NoError(t, client.Close())

	select {
	case sendErr := <-sendDone:
		require.Error(t, sendErr)
	case <-time.After(time.Second):
		t.Fatal("Send did not stop after client Close")
	}

	require.NoError(t, <-serverDone)
}
