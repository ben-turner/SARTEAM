package mutationapi

import (
	"encoding/json"
	"reflect"
)

type mutableValue struct {
	v reflect.Value
}

func (m *mutableValue) ValueToJSON() ([]byte, error) {
	return json.Marshal(m.v.Interface())
}

func (m *mutableValue) ValueFromJSON(b []byte) error {
	ptr := reflect.New(m.v.Type())
	err := json.Unmarshal(b, ptr.Interface())
	if err != nil {
		return err
	}

	v := ptr.Elem()
	if reflect.DeepEqual(m.v.Interface(), v.Interface()) {
		return ErrMutationNoChange
	}

	m.v.Set(v)

	return nil
}

func (m *mutableValue) GetField(name string) (Mutable, error) {
	return nil, ErrNotFound
}
