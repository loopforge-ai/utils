package env_test

import (
	"testing"
	"time"

	"github.com/loopforge-ai/utils/assert"
	"github.com/loopforge-ai/utils/env"
)

func Test_Get_Bool_With_FalseValue_Should_ReturnFalse(t *testing.T) {
	// Arrange
	t.Setenv("TEST_BOOL_F", "false")

	// Act
	result := env.Get("TEST_BOOL_F", true)

	// Assert
	assert.That(t, "should return false", result, false)
}

func Test_Get_Bool_With_InvalidValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_BOOL_BAD", "maybe")

	// Act
	result := env.Get("TEST_BOOL_BAD", true)

	// Assert
	assert.That(t, "should return default", result, true)
}

func Test_Get_Bool_With_TrueValue_Should_ReturnTrue(t *testing.T) {
	// Arrange
	t.Setenv("TEST_BOOL", "true")

	// Act
	result := env.Get("TEST_BOOL", false)

	// Assert
	assert.That(t, "should return true", result, true)
}

func Test_Get_Duration_With_ComplexValue_Should_ReturnParsed(t *testing.T) {
	// Arrange
	t.Setenv("TEST_DUR_COMPLEX", "2m30s")

	// Act
	result := env.Get("TEST_DUR_COMPLEX", time.Duration(0))

	// Assert
	assert.That(t, "should return 2m30s", result, 2*time.Minute+30*time.Second)
}

func Test_Get_Duration_With_InvalidValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_DUR_BAD", "forever")

	// Act
	result := env.Get("TEST_DUR_BAD", 10*time.Second)

	// Assert
	assert.That(t, "should return default", result, 10*time.Second)
}

func Test_Get_Duration_With_ValidValue_Should_ReturnParsed(t *testing.T) {
	// Arrange
	t.Setenv("TEST_DUR", "5s")

	// Act
	result := env.Get("TEST_DUR", time.Duration(0))

	// Assert
	assert.That(t, "should return 5s", result, 5*time.Second)
}

func Test_Get_Float64_With_InfValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_FLOAT_INF", "Inf")

	// Act
	result := env.Get("TEST_FLOAT_INF", 1.0)

	// Assert
	assert.That(t, "should return default for Inf", result, 1.0)
}

func Test_Get_Float64_With_NaNValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_FLOAT_NAN", "NaN")

	// Act
	result := env.Get("TEST_FLOAT_NAN", 2.0)

	// Assert
	assert.That(t, "should return default for NaN", result, 2.0)
}

func Test_Get_Float64_With_InvalidValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_FLOAT_BAD", "pi")

	// Act
	result := env.Get("TEST_FLOAT_BAD", 2.71)

	// Assert
	assert.That(t, "should return default", result, 2.71)
}

func Test_Get_Float64_With_ValidValue_Should_ReturnParsed(t *testing.T) {
	// Arrange
	t.Setenv("TEST_FLOAT", "3.14")

	// Act
	result := env.Get("TEST_FLOAT", 0.0)

	// Assert
	assert.That(t, "should return 3.14", result, 3.14)
}

func Test_Get_Int_With_EnvUnset_Should_ReturnDefault(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := env.Get("TEST_INT_MISSING", 99)

	// Assert
	assert.That(t, "should return default", result, 99)
}

func Test_Get_Int_With_InvalidValue_Should_ReturnDefault(t *testing.T) {
	// Arrange
	t.Setenv("TEST_INT_BAD", "not-a-number")

	// Act
	result := env.Get("TEST_INT_BAD", 10)

	// Assert
	assert.That(t, "should return default", result, 10)
}

func Test_Get_Int_With_ValidValue_Should_ReturnParsed(t *testing.T) {
	// Arrange
	t.Setenv("TEST_INT", "42")

	// Act
	result := env.Get("TEST_INT", 0)

	// Assert
	assert.That(t, "should return 42", result, 42)
}

func Test_Get_String_With_EmptyValue_Should_ReturnEmpty(t *testing.T) {
	// Arrange
	t.Setenv("TEST_STR_EMPTY", "")

	// Act
	result := env.Get("TEST_STR_EMPTY", "fallback")

	// Assert
	assert.That(t, "should return empty string", result, "")
}

func Test_Get_String_With_EnvSet_Should_ReturnValue(t *testing.T) {
	// Arrange
	t.Setenv("TEST_STR", "hello")

	// Act
	result := env.Get("TEST_STR", "default")

	// Assert
	assert.That(t, "should return env value", result, "hello")
}

func Test_Get_String_With_EnvUnset_Should_ReturnDefault(t *testing.T) {
	t.Parallel()
	// Arrange & Act
	result := env.Get("TEST_STR_MISSING", "fallback")

	// Assert
	assert.That(t, "should return default", result, "fallback")
}

func Test_Get_With_UnsupportedType_Should_Panic(t *testing.T) {
	// Arrange
	t.Setenv("TEST_UNSUPPORTED", "value")
	defer func() {
		// Assert
		r := recover()
		assert.That(t, "should panic", r != nil, true)
	}()

	// Act
	type custom struct{}
	env.Get("TEST_UNSUPPORTED", custom{})
}
