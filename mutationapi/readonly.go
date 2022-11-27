package mutationapi

import (
	"errors"
	"reflect"
)

var ErrReadOnly = errors.New("readonly")

type readonly struct {
	m Mutable
}

func (r *readonly) ValueToJSON() ([]byte, error) {
	return r.m.ValueToJSON()
}

func (r *readonly) ValueFromJSON(b []byte) error {
	return ErrReadOnly
}

func (r *readonly) GetField(name string) (Mutable, error) {
	field, err := r.m.GetField(name)
	if err != nil {
		return nil, err
	}

	return &readonly{field}, nil
}

func MakeReadOnly(v any) (Mutable, error) {
	m, err := mutable(reflect.ValueOf(v)) // Don't need to check for pointer here because we aren't modifying the value
	if err != nil {
		return nil, err
	}

	return &readonly{m}, nil
}
