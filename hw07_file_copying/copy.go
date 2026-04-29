package main

import (
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/TommyBlackrock/progressbar" //nolint:depguard
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	ErrNegativeOffset        = errors.New("negative offset is not allowed")
	ErrNegativeLimit         = errors.New("negative limit is not allowed")
	ErrSameFile              = errors.New("source and destination are the same file")
	ErrSourceChanged         = errors.New("source file changed during copying")
	ErrVerificationFailed    = errors.New("copied data verification failed")
)

const copyBufferSize = 1024 * 1024

type progressWriter struct {
	total         int64
	copied        int64
	taskName      string
	bar           *progressbar.ConsoleProgressBar
	progressMuted bool
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	if w.total > 0 && !w.progressMuted {
		w.copied += int64(n)
		percent := int(w.copied * 100 / w.total)
		if percent >= 100 {
			percent = 99
		}
		if err := w.bar.Update(w.taskName, percent); err != nil {
			// Progress rendering must not break data copying.
			w.progressMuted = true
		}
	}

	return n, nil
}

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
	srcInitialModTime := srcInfo.ModTime()

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

	bytesToCopy := calcBytesToCopy(srcSize, offset, limit)

	dstDir := filepath.Dir(toPath)
	tmpFile, err := os.CreateTemp(dstDir, filepath.Base(toPath)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary destination file: %w", err)
	}
	tmpPath := tmpFile.Name()
	renameDone := false
	defer func() {
		if !renameDone {
			_ = os.Remove(tmpPath)
		}
	}()

	defer func() {
		_ = tmpFile.Close()
	}()

	srcRange := io.NewSectionReader(src, offset, bytesToCopy)
	srcHash := crc32.NewIEEE()

	bar := progressbar.NewConsoleProgressBar(25)
	taskName := "copy"
	bar.Add(taskName)

	progress := &progressWriter{
		total:    bytesToCopy,
		taskName: taskName,
		bar:      bar,
	}

	copyReader := io.TeeReader(srcRange, io.MultiWriter(progress, srcHash))
	buffer := make([]byte, copyBufferSize)
	copied, err := io.CopyBuffer(tmpFile, copyReader, buffer)
	if err != nil {
		return fmt.Errorf("copy data: %w", err)
	}

	if copied != bytesToCopy {
		return fmt.Errorf("copied %d bytes, expected to copy %d bytes", copied, bytesToCopy)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync destination temporary file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close destination temporary file: %w", err)
	}

	tmpInfo, err := os.Stat(tmpPath)
	if err != nil {
		return fmt.Errorf("stat destination temporary file: %w", err)
	}
	if tmpInfo.Size() != bytesToCopy {
		return fmt.Errorf(
			"%w: destination size mismatch: got %d bytes, expected %d bytes",
			ErrVerificationFailed,
			tmpInfo.Size(),
			bytesToCopy,
		)
	}

	srcInfoAfter, err := os.Stat(fromPath)
	if err != nil {
		return fmt.Errorf("stat source file after copy: %w", err)
	}
	if srcInfoAfter.Size() != srcSize || !sameModTime(srcInfoAfter.ModTime(), srcInitialModTime) {
		return fmt.Errorf(
			"%w: before(size=%d modtime=%s), after(size=%d modtime=%s)",
			ErrSourceChanged,
			srcSize,
			srcInitialModTime.Format(time.RFC3339Nano),
			srcInfoAfter.Size(),
			srcInfoAfter.ModTime().Format(time.RFC3339Nano),
		)
	}

	dstHash, err := crc32File(tmpPath)
	if err != nil {
		return fmt.Errorf("calculate destination crc32: %w", err)
	}
	if srcHash.Sum32() != dstHash {
		return fmt.Errorf(
			"%w: source crc32=%08x destination crc32=%08x",
			ErrVerificationFailed,
			srcHash.Sum32(),
			dstHash,
		)
	}

	_ = bar.Finish(taskName)

	if err := os.Rename(tmpPath, toPath); err != nil {
		return fmt.Errorf("rename temporary file to destination: %w", err)
	}
	renameDone = true

	return nil
}

func calcBytesToCopy(srcSize, offset, limit int64) int64 {
	if offset >= srcSize {
		return 0
	}

	bytesToCopy := srcSize - offset
	if limit > 0 && limit < bytesToCopy {
		bytesToCopy = limit
	}

	return bytesToCopy
}

func crc32File(path string) (uint32, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	hash := crc32.NewIEEE()
	buffer := make([]byte, copyBufferSize)
	if _, err := io.CopyBuffer(hash, f, buffer); err != nil {
		return 0, err
	}

	return hash.Sum32(), nil
}

func sameModTime(left, right time.Time) bool {
	return left.Equal(right)
}
