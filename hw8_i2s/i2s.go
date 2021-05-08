package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer")
	}
	if rv.IsNil() {
		return fmt.Errorf("is nil")
	}

	rv = rv.Elem()

	switch rv.Type().Kind() {
	case reflect.Struct:
		dict, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("not dict")
		}
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i).Name

			value, ok := dict[field]
			if !ok {
				return fmt.Errorf("field not found: %s", field)
			}

			if err := i2s(value, rv.Field(i).Addr().Interface()); err != nil {
				return fmt.Errorf("failed to parse field %s: %s", field, err)
			}
		}
	case reflect.Slice:
		arr, ok := data.([]interface{})
		if !ok {
			return fmt.Errorf("not an slice")
		}
		for idx, el := range arr {
			elValue := reflect.New(rv.Type().Elem())
			if err := i2s(el, elValue.Interface()); err != nil {
				return fmt.Errorf("failed to parse slice element %d: %s", idx, err)
			}

			rv.Set(reflect.Append(rv, elValue.Elem()))
		}
	case reflect.Int:
		f, ok := data.(float64)
		if !ok {
			return fmt.Errorf("not a float64")
		}

		rv.SetInt(int64(f))

	case reflect.String:
		str, ok := data.(string)
		if !ok {
			return fmt.Errorf("not a string")
		}

		rv.SetString(str)

	case reflect.Bool:
		b, ok := data.(bool)
		if !ok {
			return fmt.Errorf("not a bool")
		}

		rv.SetBool(b)

	default:
		return fmt.Errorf("type %s is not supported", rv.Kind())

	}
	return nil
}
