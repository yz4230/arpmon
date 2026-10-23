package testutil

import (
	"reflect"
	"testing"
)

func AssertAllFieldsNonZero(t *testing.T, v any) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			t.Fatalf("expected non-nil pointer, got nil")
		}
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		t.Fatalf("expected struct, got %s", rv.Kind())
	}

	issues := []string{}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		if field.IsZero() {
			fieldType := rt.Field(i)
			name := fieldType.Name
			issues = append(issues, name)
		}
	}

	if len(issues) > 0 {
		if len(issues) == 1 {
			t.Fatalf("field %s is zero value", issues[0])
		}
		t.Fatalf("fields %v are zero value", issues)
	}
}
