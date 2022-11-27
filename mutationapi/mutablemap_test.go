package mutationapi

import (
	"reflect"
	"testing"
)

func TestIsValidMapKey(t *testing.T) {
	type test struct {
		kind reflect.Kind
		want bool
	}

	tests := []test{
		{reflect.Invalid, false},
		{reflect.Bool, false},
		{reflect.Int, true},
		{reflect.Int8, true},
		{reflect.Int16, true},
		{reflect.Int32, true},
		{reflect.Int64, true},
		{reflect.Uint, true},
		{reflect.Uint8, true},
		{reflect.Uint16, true},
		{reflect.Uint32, true},
		{reflect.Uint64, true},
		{reflect.Uintptr, false},
		{reflect.Float32, true},
		{reflect.Float64, true},
		{reflect.Complex64, false},
		{reflect.Complex128, false},
		{reflect.Array, false},
		{reflect.Chan, false},
		{reflect.Func, false},
		{reflect.Interface, false},
		{reflect.Map, false},
		{reflect.Pointer, false},
		{reflect.Slice, false},
		{reflect.String, true},
		{reflect.Struct, false},
		{reflect.UnsafePointer, false},
	}

	for _, test := range tests {
		test := test // copy so we can use t.Parallel

		t.Run(test.kind.String(), func(t *testing.T) {
			t.Parallel()

			got := isValidMapKey(test.kind)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}
