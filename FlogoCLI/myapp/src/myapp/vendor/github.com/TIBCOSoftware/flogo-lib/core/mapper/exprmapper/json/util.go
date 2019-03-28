package json

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func ToArray(val interface{}) ([]interface{}, error) {

	switch t := val.(type) {
	case []interface{}:
		return t, nil

	case []map[string]interface{}:
		var a []interface{}
		for _, v := range t {
			a = append(a, v)
		}
		return a, nil
	case string:
		a := make([]interface{}, 0)
		if t != "" {
			err := json.Unmarshal([]byte(t), &a)
			if err != nil {
				return nil, fmt.Errorf("unable to coerce %#v to map[string]interface{}", val)
			}
		}
		return a, nil
	case nil:
		return nil, nil
	default:
		s := reflect.ValueOf(val)
		if s.Kind() == reflect.Slice {
			a := make([]interface{}, s.Len())

			for i := 0; i < s.Len(); i++ {
				a[i] = s.Index(i).Interface()
			}
			return a, nil
		}
		return nil, fmt.Errorf("unable to coerce %#v to []interface{}", val)
	}
}

func GetFieldByName(object interface{}, name string) (reflect.Value, error) {
	val := reflect.ValueOf(object)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	} else {
	}

	field := val.FieldByName(name)
	if field.IsValid() {
		return field, nil
	}

	typ := reflect.TypeOf(object)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	for i := 0; i < typ.NumField(); i++ {
		p := typ.Field(i)
		if !p.Anonymous {
			if p.Tag != "" && len(p.Tag) > 0 {
				if name == p.Tag.Get("json") {
					return val.FieldByName(typ.Field(i).Name), nil
				}
			}
		}
	}

	return reflect.Value{}, nil
}

func IsMapperableType(data interface{}) bool {
	switch t := data.(type) {
	case map[string]interface{}, map[string]string, []int, []int64, []string, []map[string]interface{}, []map[string]string:
		return true
	case []interface{}:
		//
		isStruct := true
		for _, v := range t {
			if v != nil {
				if reflect.TypeOf(v).Kind() == reflect.Struct {
					isStruct = false
					break
				} else if reflect.TypeOf(v).Kind() == reflect.Ptr {
					isStruct = false
					break
				}
			}
		}
		return isStruct
	}

	return false
}
