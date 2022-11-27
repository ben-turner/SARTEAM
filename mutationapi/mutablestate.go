package mutationapi

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Mutable interface {
	ValueToJSON() ([]byte, error)
	ValueFromJSON([]byte) error

	GetField(string) (Mutable, error)
}

type MutableState struct {
	data Mutable
}

func NewMutableState(value any) (*MutableState, error) {
	v := reflect.ValueOf(value)

	if v.Type().Kind() != reflect.Ptr {
		return nil, errors.New("value must be a pointer")
	}

	data, err := mutable(v)
	if err != nil {
		return nil, err
	}

	return &MutableState{
		data: data,
	}, nil
}

func (m *MutableState) ApplyMutation(mut *Mutation) error {
	value := m.data
	for _, path := range mut.Path {
		field, err := value.GetField(path)
		if err != nil {
			return err
		}
		value = field
	}

	switch mut.Action {
	case MutationActionRead:
		if mut.Conn == nil {
			return errors.New("invalid connection")
		}

		res, err := value.ValueToJSON()
		if err != nil {
			return err
		}

		reply := &Mutation{
			ClientID:  mut.ClientID,
			Action:    MutationActionUpdate,
			Path:      mut.Path,
			Body:      res,
			Timestamp: time.Now(),
		}

		return mut.Conn.Send(reply)
	case MutationActionUpdate:
		return value.ValueFromJSON(mut.BodyAsBytes())
	default:
		return errors.New("invalid action")
	}
}

func mutableFromArrayOrSlice(v reflect.Value) (Mutable, error) {
	return nil, nil
}

func mutable(v reflect.Value) (Mutable, error) {
	ptr := v
	for ptr.Kind() == reflect.Ptr || ptr.Kind() == reflect.Interface {
		if ptr.IsNil() {
			ptr.Set(reflect.New(ptr.Type().Elem()))
		}

		ptr = ptr.Elem()
	}

	direct, ok := v.Interface().(Mutable)
	if ok {
		return direct, nil
	}

	switch ptr.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64, reflect.String, reflect.Bool:
		return &mutableValue{ptr}, nil
	case reflect.Array, reflect.Slice:
		return mutableFromArrayOrSlice(ptr)
	case reflect.Struct:
		return mutableFromStruct(ptr)
	case reflect.Map:
		return mutableFromMap(ptr)
	default:
		return nil, fmt.Errorf("unsupported type: %v", v.Kind())
	}
}

func MakeMutable(value any) (Mutable, error) {
	v := reflect.ValueOf(value)

	if v.Type().Kind() != reflect.Ptr {
		return nil, errors.New("value must be a pointer")
	}

	return mutable(v)
}
