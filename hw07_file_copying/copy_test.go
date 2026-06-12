package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const inputFile = "testdata/input.txt"

func TestCopy_Success(t *testing.T) {
	tests := []struct {
		name   string
		offset int64
		limit  int64
		want   string
	}{
		{"offset0_limit0", 0, 0, "testdata/out_offset0_limit0.txt"},
		{"offset0_limit10", 0, 10, "testdata/out_offset0_limit10.txt"},
		{"offset0_limit1000", 0, 1000, "testdata/out_offset0_limit1000.txt"},
		{"offset0_limit10000", 0, 10000, "testdata/out_offset0_limit10000.txt"},
		{"offset100_limit1000", 100, 1000, "testdata/out_offset100_limit1000.txt"},
		{"offset6000_limit1000", 6000, 1000, "testdata/out_offset6000_limit1000.txt"},
		{"limit_larger_than_file", 0, 100000, "testdata/out_offset0_limit0.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "out.txt")

			err := Copy(inputFile, tmp, tt.offset, tt.limit)
			require.NoError(t, err)

			wantData, err := os.ReadFile(tt.want)
			require.NoError(t, err)
			gotData, err := os.ReadFile(tmp)
			require.NoError(t, err)

			assert.Equal(t, wantData, gotData)
		})
	}
}

func TestCopy_OffsetEqualsFileSize(t *testing.T) {
	srcInfo, err := os.Stat(inputFile)
	require.NoError(t, err)

	tmp := filepath.Join(t.TempDir(), "out.txt")
	err = Copy(inputFile, tmp, srcInfo.Size(), 0)
	require.NoError(t, err)

	dstInfo, err := os.Stat(tmp)
	require.NoError(t, err)
	assert.Equal(t, int64(0), dstInfo.Size())
}

func TestCopy_SameFile(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "same.txt")
	initial := []byte("sample-content-for-same-file-check")
	require.NoError(t, os.WriteFile(filePath, initial, 0o644))

	err := Copy(filePath, filePath, 0, 10)
	require.ErrorIs(t, err, ErrSameFile)

	got, readErr := os.ReadFile(filePath)
	require.NoError(t, readErr)
	assert.Equal(t, initial, got)
}

func TestCopy_Errors(t *testing.T) {
	srcInfo, err := os.Stat(inputFile)
	require.NoError(t, err)

	tests := []struct {
		name      string
		from      string
		offset    int64
		limit     int64
		wantError error
	}{
		{"negative offset", inputFile, -1, 100, ErrNegativeOffset},
		{"negative limit", inputFile, 0, -100, ErrNegativeLimit},
		{"offset exceeds file size", inputFile, srcInfo.Size() + 1, 100, ErrOffsetExceedsFileSize},
		{"source file not exists", "testdata/not_exists.txt", 0, 0, os.ErrNotExist},
		{"source is directory", "testdata", 0, 0, ErrUnsupportedFile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := filepath.Join(t.TempDir(), "out.txt")
			err := Copy(tt.from, tmp, tt.offset, tt.limit)

			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}
