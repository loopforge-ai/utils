package env

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"
)

// Get retrieves an environment variable and parses it to the type of the default value.
// If the value is not set or cannot be parsed, the default value is returned.
// Supported types: bool, int, float64, string, time.Duration.
func Get[T any](key string, def T) T {
	value, ok := os.LookupEnv(key)
	if !ok {
		return def
	}
	result, ok := parse[T](value, def)
	if !ok {
		return def
	}
	return result
}

func parse[T any](value string, def T) (T, bool) {
	var result any
	var err error

	switch any(def).(type) {
	case bool:
		result, err = strconv.ParseBool(value)
	case float64:
		result, err = parseFloat64(value)
		if err != nil {
			return def, false
		}
	case int:
		result, err = strconv.Atoi(value)
	case string:
		return any(value).(T), true
	case time.Duration:
		result, err = time.ParseDuration(value)
	default:
		panic(fmt.Sprintf("env.Get: unsupported type %T", def))
	}

	if err != nil {
		return def, false
	}
	return result.(T), true
}

func parseFloat64(value string) (float64, error) {
	f, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if math.IsInf(f, 0) || math.IsNaN(f) {
		return 0, fmt.Errorf("non-finite float: %s", value)
	}
	return f, nil
}
