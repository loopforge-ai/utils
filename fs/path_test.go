package fs_test

import (
	"testing"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/fs"
)

func Test_IsSafePath_With_AbsolutePath_Should_ReturnFalse(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := fs.IsSafePath("/etc/passwd")

	// Assert
	assert.That(t, "absolute path should be unsafe", result, false)
}

func Test_IsSafePath_With_ParentTraversal_Should_ReturnFalse(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := fs.IsSafePath("../secret")

	// Assert
	assert.That(t, "parent traversal should be unsafe", result, false)
}

func Test_IsSafePath_With_NestedTraversal_Should_ReturnFalse(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := fs.IsSafePath("foo/../../etc/passwd")

	// Assert
	assert.That(t, "nested traversal should be unsafe", result, false)
}

func Test_IsSafePath_With_RelativePath_Should_ReturnTrue(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := fs.IsSafePath("pkg/main.go")

	// Assert
	assert.That(t, "relative path should be safe", result, true)
}

func Test_IsSafePath_With_SingleFile_Should_ReturnTrue(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := fs.IsSafePath("main.go")

	// Assert
	assert.That(t, "single file should be safe", result, true)
}
