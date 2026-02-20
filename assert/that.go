// Package assert provides test assertion helpers.
package assert

import (
	"reflect"
	"testing"
)

// That fails the test with a descriptive message if got does not deeply equal expected.
func That(t *testing.T, desc string, got, expected any) {
	t.Helper()
	if isNil(got) && isNil(expected) {
		return
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("%s: got %v (%T), expected %v (%T)", desc, got, got, expected, expected)
	}
}

// isNil returns true if v is nil or a typed nil (e.g. (*T)(nil)).
func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() { //nolint:exhaustive // only nillable kinds checked
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	}
	return false
}
