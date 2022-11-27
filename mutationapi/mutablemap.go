package mutationapi

import (
	"encoding/json"
	"errors"
	"reflect"
)

func isValidMapKey(k reflect.Kind) bool {
	return k == reflect.String || // String
		(k >= reflect.Int && k <= reflect.Uint64) || // Integers, excluding uintptr
		k == reflect.Float32 || k == reflect.Float64 // Floats
}

type mutableMap struct {
	v     reflect.Value
	cache map[string]Mutable
}

func (m *mutableMap) ValueToJSON() ([]byte, error) {
	return json.Marshal(m.v.Interface())
}

func (m *mutableMap) ValueFromJSON(b []byte) error {
	v := reflect.MakeMap(m.v.Type())

	err := json.Unmarshal(b, v.Interface())
	if err != nil {
		return err
	}

	if reflect.DeepEqual(m.v.Interface(), v.Interface()) {
		return ErrMutationNoChange
	}

	m.v.Set(v)
	return nil
}

func (m *mutableMap) GetField(name string) (Mutable, error) {
	cached, ok := m.cache[name]
	if ok {
		return cached, nil
	}

	field := m.v.MapIndex(reflect.ValueOf(name))

	if !field.IsValid() {
		field = reflect.New(m.v.Type().Elem()).Elem()
		m.v.SetMapIndex(reflect.ValueOf(name), field)
	}

	res, err := mutable(field)
	if err != nil {
		return nil, err
	}

	m.cache[name] = res

	return res, nil
}

func mutableFromMap(v reflect.Value) (Mutable, error) {
	keyKind := v.Type().Key().Kind()

	if !isValidMapKey(keyKind) {
		return nil, errors.New("map key must be of kind string")
	}

	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	return &mutableMap{
		v:     v,
		cache: make(map[string]Mutable),
	}, nil
}
