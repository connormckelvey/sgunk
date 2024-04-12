package util

import (
	"errors"
	"reflect"

	"github.com/mitchellh/mapstructure"
)

type Marshaler interface {
	MarshalMap() (map[string]any, error)
}

func Unmarshal(data map[string]any, output any) error {
	return mapstructure.Decode(data, output)
}

func MarshalMap(v any) (map[string]any, error) {
	if m, ok := v.(Marshaler); ok {
		return m.MarshalMap()
	}

	rt := reflect.TypeOf(v)
	switch rt.Kind() {
	case reflect.Struct:
		return marshalStruct(reflect.ValueOf(v))
	case reflect.Pointer:
		return MarshalMap(reflect.ValueOf(v).Elem().Interface())
	case reflect.Map:
		if rt.Key().Kind() != reflect.String {
			return marshalMap(reflect.ValueOf(rt))
		}
	}
	return nil, errors.New("input must be a struct, struct pointer, map[string]any, or a MapMarshaler")
}

var marshalerType = reflect.TypeOf((*Marshaler)(nil)).Elem()

func marshalValue(v reflect.Value) (any, error) {
	rt := v.Type()
	if rt.Implements(marshalerType) {
		return v.Interface().(Marshaler).MarshalMap()
	}

	switch rt.Kind() {
	case reflect.Struct:
		return marshalStruct(v)
	case reflect.Map:
		return marshalMap(v)
	case reflect.Slice:
		return marshalSlice(v)
	case reflect.Pointer:
		return marshalValue(v.Elem())
	default:
		if v.CanInterface() {
			return v.Interface(), nil
		}
		return nil, nil
	}
}

func marshalStruct(v reflect.Value) (map[string]any, error) {
	rt := v.Type()

	m := make(map[string]any)
	for i := 0; i < rt.NumField(); i++ {
		rf := rt.Field(i)
		key, ok := rf.Tag.Lookup("mapstructure")
		if !ok {
			key = rf.Name
		}
		fv := v.Field(i)
		v, err := marshalValue(fv)
		if err != nil {
			return nil, err
		}
		m[key] = v
	}
	return m, nil
}

func marshalMap(v reflect.Value) (map[string]any, error) {
	keys := v.MapKeys()
	m := make(map[string]any)
	for _, key := range keys {
		if key.Type().Kind() != reflect.String {
			return nil, nil
		}
		v, err := marshalValue(v.MapIndex(key))
		if err != nil {
			return nil, err
		}
		m[key.Interface().(string)] = v
	}
	return m, nil
}

func marshalSlice(v reflect.Value) ([]any, error) {
	var ms []any
	for i := 0; i < v.Len(); i++ {
		value, err := marshalValue(v.Index(i))
		if err != nil {
			return nil, err
		}
		ms = append(ms, value)
	}
	return ms, nil
}
