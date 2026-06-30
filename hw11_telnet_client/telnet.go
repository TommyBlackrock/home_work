package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var ErrNotConnected = errors.New("telnet client is not connected")

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer

	mu          sync.Mutex
	conn        net.Conn
	inputClosed bool
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

func (c *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("connect to %s: %w", c.address, err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	return nil
}

func (c *telnetClient) Close() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	in := c.in
	closeInput := !c.inputClosed
	c.inputClosed = true
	c.mu.Unlock()

	var closeErrors []error
	if conn != nil {
		if err := conn.Close(); err != nil && !isClosedResourceError(err) {
			closeErrors = append(closeErrors, fmt.Errorf("close connection: %w", err))
		}
	}

	// Closing input interrupts Send if it is blocked while reading STDIN.
	if closeInput && in != nil {
		if err := in.Close(); err != nil && !isClosedResourceError(err) {
			closeErrors = append(closeErrors, fmt.Errorf("close input: %w", err))
		}
	}

	return errors.Join(closeErrors...)
}

func (c *telnetClient) CloseWrite() error {
	conn, err := c.connection()
	if err != nil {
		return err
	}

	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return c.Close()
	}

	if err := tcpConn.CloseWrite(); err != nil {
		return fmt.Errorf("close write side: %w", err)
	}

	return nil
}

func (c *telnetClient) Send() error {
	conn, err := c.connection()
	if err != nil {
		return err
	}

	if _, err := io.Copy(conn, c.in); err != nil {
		return fmt.Errorf("send data: %w", err)
	}

	return nil
}

func (c *telnetClient) Receive() error {
	conn, err := c.connection()
	if err != nil {
		return err
	}

	if _, err := io.Copy(c.out, conn); err != nil {
		return fmt.Errorf("receive data: %w", err)
	}

	return nil
}

func (c *telnetClient) connection() (net.Conn, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil, ErrNotConnected
	}

	return c.conn, nil
}

func isClosedResourceError(err error) bool {
	return errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, os.ErrClosed)
}
