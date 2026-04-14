package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrNegativeOffset        = errors.New("negative offset is not allowed")
	ErrNegativeLimit         = errors.New("negative limit is not allowed")
	ErrSameFile              = errors.New("source and destination are the same file")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	if offset < 0 {
		return ErrNegativeOffset
	}
	if limit < 0 {
		return ErrNegativeLimit
	}

	src, err := os.OpenFile(fromPath, os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer src.Close()

	srcInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("stat source file: %w", err)
	}
	if !srcInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	srcSize := srcInfo.Size()
	if offset > srcSize {
		return ErrOffsetExceedsFileSize
	}

	dstInfo, err := os.Stat(toPath)
	switch {
	case err == nil:
		if os.SameFile(srcInfo, dstInfo) {
			return ErrSameFile
		}
	case errors.Is(err, os.ErrNotExist):
		// Destination file does not exist yet: this is a valid case.
	default:
		return fmt.Errorf("stat destination file: %w", err)
	}

	bytesToCopy := srcSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	dst, err := os.OpenFile(toPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := src.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("seek to offset %d: %w", offset, err)
	}

	bar := pb.New64(bytesToCopy)
	bar.Start()
	defer bar.Finish()

	progressReader := bar.NewProxyReader(src)

	copied, err := io.CopyN(dst, progressReader, bytesToCopy)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy data: %w", err)
	}

	if copied != bytesToCopy {
		return fmt.Errorf("copied %d bytes, expected to copy %d bytes", copied, bytesToCopy)
	}

	return nil
}
