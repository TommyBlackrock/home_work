package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"time"
)

const defaultTimeout = 10 * time.Second

type operationResult struct {
	name string
	err  error
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, in io.ReadCloser, out, errOut io.Writer) error {
	timeout, address, err := parseArgs(args, errOut)
	if err != nil {
		return err
	}

	client := NewTelnetClient(address, timeout, in, out)
	if err := client.Connect(); err != nil {
		return err
	}

	fmt.Fprintf(errOut, "...Connected to %s\n", address)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	done := make(chan operationResult, 2)
	go func() {
		done <- operationResult{name: "send", err: client.Send()}
	}()
	go func() {
		done <- operationResult{name: "receive", err: client.Receive()}
	}()

	select {
	case <-ctx.Done():
		fmt.Fprintln(errOut, "...Interrupted")
	case result := <-done:
		printOperationResult(errOut, result)
		if result.err != nil && !isExpectedCloseError(result.err) {
			return closeClient(client, result.err)
		}
		if result.name == "send" {
			if err := closeWrite(client); err != nil && !isExpectedCloseError(err) {
				return closeClient(client, err)
			}
			if err := waitReceiveAfterEOF(ctx, errOut, done); err != nil {
				return closeClient(client, err)
			}
		}
	}

	if err := client.Close(); err != nil && !isExpectedCloseError(err) {
		return err
	}

	waitSecondOperation(done)
	return nil
}

func parseArgs(args []string, errOut io.Writer) (time.Duration, string, error) {
	flags := flag.NewFlagSet("go-telnet", flag.ContinueOnError)
	flags.SetOutput(errOut)

	var timeout time.Duration
	flags.DurationVar(&timeout, "timeout", defaultTimeout, "connection timeout")
	if err := flags.Parse(args); err != nil {
		return 0, "", err
	}

	if flags.NArg() != 2 {
		return 0, "", errors.New("usage: go-telnet [--timeout=10s] host port")
	}

	return timeout, net.JoinHostPort(flags.Arg(0), flags.Arg(1)), nil
}

func printOperationResult(out io.Writer, result operationResult) {
	switch result.name {
	case "send":
		fmt.Fprintln(out, "...EOF")
	case "receive":
		fmt.Fprintln(out, "...Connection was closed by peer")
	}
}

func waitSecondOperation(done <-chan operationResult) {
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}
}

func waitReceiveAfterEOF(ctx context.Context, errOut io.Writer, done <-chan operationResult) error {
	select {
	case <-ctx.Done():
		fmt.Fprintln(errOut, "...Interrupted")
	case result := <-done:
		if result.name == "receive" {
			printOperationResult(errOut, result)
			if result.err != nil && !isExpectedCloseError(result.err) {
				return result.err
			}
		}
	case <-time.After(500 * time.Millisecond):
	}

	return nil
}

func closeWrite(client TelnetClient) error {
	writeCloser, ok := client.(interface {
		CloseWrite() error
	})
	if !ok {
		return client.Close()
	}

	return writeCloser.CloseWrite()
}

func isExpectedCloseError(err error) bool {
	return errors.Is(err, io.EOF) ||
		errors.Is(err, io.ErrClosedPipe) ||
		errors.Is(err, net.ErrClosed) ||
		errors.Is(err, os.ErrClosed)
}

func closeClient(client TelnetClient, operationErr error) error {
	closeErr := client.Close()
	if closeErr == nil || isExpectedCloseError(closeErr) {
		return operationErr
	}

	return errors.Join(operationErr, closeErr)
}
