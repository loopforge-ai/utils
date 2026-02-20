package assert_test

import (
	"testing"

	"github.com/loopforge-ai/utils/assert"
)

func Test_That_With_EqualSlices_Should_Pass(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &testing.T{}

	// Act
	assert.That(mock, "slices match", []int{1, 2, 3}, []int{1, 2, 3})

	// Assert
	if mock.Failed() {
		t.Error("expected test to pass for equal slices")
	}
}

func Test_That_With_EqualStrings_Should_Pass(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &testing.T{}

	// Act
	assert.That(mock, "strings match", "hello", "hello")

	// Assert
	if mock.Failed() {
		t.Error("expected test to pass for equal strings")
	}
}

func Test_That_With_EqualValues_Should_Pass(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &testing.T{}

	// Act
	assert.That(mock, "should equal", 42, 42)

	// Assert — mock.Failed() would be true if assertion failed
	if mock.Failed() {
		t.Error("expected test to pass for equal values")
	}
}

func Test_That_With_NilValues_Should_Pass(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &testing.T{}

	// Act
	assert.That(mock, "both nil", nil, nil)

	// Assert
	if mock.Failed() {
		t.Error("expected test to pass for nil == nil")
	}
}

func Test_That_With_UnequalValues_Should_Fail(t *testing.T) {
	t.Parallel()
	// Arrange
	mock := &testing.T{}

	// Act
	assert.That(mock, "should differ", 1, 2)

	// Assert
	if !mock.Failed() {
		t.Error("expected test to fail for unequal values")
	}
}
