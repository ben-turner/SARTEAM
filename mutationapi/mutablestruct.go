package mutationapi

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type mutableStruct struct {
	v      reflect.Value
	fields map[string]int
	cache  map[string]Mutable
}

func (m *mutableStruct) ValueToJSON() ([]byte, error) {
	return json.Marshal(m.v.Interface())
}

func (m *mutableStruct) ValueFromJSON(b []byte) error {
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

func (m *mutableStruct) GetField(name string) (Mutable, error) {
	cached, ok := m.cache[name]
	if ok {
		return cached, nil
	}

	fieldNum, ok := m.fields[name]
	if !ok {
		return nil, fmt.Errorf("invalid field: %q", name)
	}

	field := m.v.Field(fieldNum)

	direct, ok := field.Interface().(Mutable) // Implements Mutable interface
	if ok {
		m.cache[name] = direct
		return direct, nil
	}

	res, err := mutable(field)
	if err != nil {
		return nil, err
	}

	m.cache[name] = res

	return res, nil
}

func mutableFromStruct(v reflect.Value) (Mutable, error) {
	numFields := v.NumField()
	t := v.Type()

	fields := make(map[string]int, numFields)
	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		name := field.Tag.Get("mutationapi")
		if name == "" {
			name = field.Name
		}
		fields[name] = i
	}

	return &mutableStruct{
		v:      v,
		fields: fields,
		cache:  make(map[string]Mutable),
	}, nil
}
