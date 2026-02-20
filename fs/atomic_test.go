package fs_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/fs"
)

func Test_AtomicWrite_With_EmptyContent_Should_CreateEmptyFile(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")

	// Act
	err := fs.AtomicWrite(path, []byte{})

	// Assert
	assert.That(t, "error should be nil", err, nil)
	info, statErr := os.Stat(path)
	assert.That(t, "file should exist", statErr, nil)
	assert.That(t, "file should be empty", info.Size(), int64(0))
}

func Test_AtomicWrite_With_ExistingFile_Should_Overwrite(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	_ = os.WriteFile(path, []byte("old"), 0o600)

	// Act
	err := fs.AtomicWrite(path, []byte("new"))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	data, _ := os.ReadFile(path) //nolint:gosec // test path
	assert.That(t, "content should be overwritten", string(data), "new")
}

func Test_AtomicWrite_With_FilePermissions_Should_Set0600(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "perm.txt")

	// Act
	err := fs.AtomicWrite(path, []byte("data"))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	info, _ := os.Stat(path)
	assert.That(t, "permissions should be 0600", info.Mode().Perm(), os.FileMode(0o600))
}

func Test_AtomicWrite_With_ValidPath_Should_CreateFile(t *testing.T) {
	t.Parallel()
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")

	// Act
	err := fs.AtomicWrite(path, []byte("hello"))

	// Assert
	assert.That(t, "error should be nil", err, nil)
	data, readErr := os.ReadFile(path) //nolint:gosec // test path
	assert.That(t, "read error should be nil", readErr, nil)
	assert.That(t, "content should match", string(data), "hello")
}
