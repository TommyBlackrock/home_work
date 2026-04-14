package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	from, to      string
	limit, offset int64
)

func init() {
	flag.StringVar(&from, "from", "", "file to read from")
	flag.StringVar(&to, "to", "", "file to write to")
	flag.Int64Var(&limit, "limit", 0, "limit of bytes to copy")
	flag.Int64Var(&offset, "offset", 0, "offset in input file")
}

func main() {
	flag.Parse()
	if from == "" {
		fmt.Println("Error: 'from' flag is required")
		os.Exit(1)
	}
	if to == "" {
		fmt.Println("Error: 'to' flag is required")
		os.Exit(1)
	}

	err := Copy(from, to, offset, limit)
	if err == nil {
		fmt.Printf("File %s copied successfully to %s\n", from, to)
		return
	}
	switch {
	case errors.Is(err, ErrNegativeOffset):
		fmt.Fprintln(os.Stderr, "Error: offset cannot be negative")
	case errors.Is(err, ErrNegativeLimit):
		fmt.Fprintln(os.Stderr, "Error: limit cannot be negative")
	case errors.Is(err, ErrOffsetExceedsFileSize):
		fmt.Fprintln(os.Stderr, "Error: offset exceeds file size")
	case errors.Is(err, ErrSameFile):
		fmt.Fprintln(os.Stderr, "Error: source and destination must be different files")
	case errors.Is(err, ErrUnsupportedFile):
		fmt.Fprintln(os.Stderr, "Error: only regular files are supported")
	case errors.Is(err, os.ErrNotExist):
		fmt.Fprintf(os.Stderr, "Error: source file %s does not exist\n", from)
	case errors.Is(err, os.ErrPermission):
		fmt.Fprintf(os.Stderr, "Error: permission denied (reading %s or writing %s)\n", from, to)
	default:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	os.Exit(1)
}
