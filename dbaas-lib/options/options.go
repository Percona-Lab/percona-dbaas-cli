package options

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Parse parses options from the given string in format "object.paramValue=val,objectTwo.paramValue=val"
// and assigned it into the given "to" of type type
func Parse(to interface{}, typ reflect.Type, options string) error {
	if options == "" {
		return nil
	}
	options = strings.ToLower(options)
	opts := make(map[string]string)
	validConfKeys(typ, opts, "", "")
	optArr := strings.Split(options, ",")
	for _, str := range optArr {
		v := strings.Split(str, "=")
		if _, ok := opts[v[0]]; !ok {
			return errors.Errorf("invalidd option %s", v[0])
		}
		if len(v) > 1 {
			fs := strings.Split(opts[v[0]], ".")
			rv := reflect.ValueOf(to).Elem()
			for _, f := range fs {
				rv = rv.FieldByName(f)
			}
			err := setValue(rv, v[1])
			if err != nil {
				return errors.Wrapf(err, "set value %s=%s", v[0], v[1])
			}
		}
	}

	return nil
}

func setValue(val reflect.Value, value string) error {

	if val.Kind() == reflect.Ptr {
		if val.IsZero() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		val = reflect.Indirect(val)
	}

	switch val.Kind() {
	default:
		// TODO: maps, slices
		return errors.Errorf("type %v not implemented", val.Kind())
	case reflect.Map:
		v, err := parseMapValue(value, val)
		if err != nil {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.Set(v)
	case reflect.Slice:
		v, err := parseSliceValue(value, val)
		if err != nil {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.Set(v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || val.OverflowInt(v) {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil || val.OverflowUint(v) {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.SetUint(v)
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil || val.OverflowFloat(v) {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return errors.Errorf("parse value %s: %v", val, err)
		}
		val.SetBool(v)
	case reflect.String:
		val.SetString(value)

	}
	return nil
}

func parseMapValue(s string, refValue reflect.Value) (reflect.Value, error) {
	value := reflect.MakeMap(refValue.Type())

	sSlice := strings.Split(s, ";")
	if len(sSlice) == 0 {
		return value, errors.New("empty value")
	}

	keyValue := reflect.Indirect(reflect.New(value.Type().Key()))
	mapValue := reflect.Indirect(reflect.New(value.Type().Elem()))
	for _, v := range sSlice {
		vSlice := strings.Split(v, ":")
		if len(vSlice) != 2 {
			return value, errors.New("empty map value")
		}
		setValue(keyValue, vSlice[0])
		setValue(mapValue, vSlice[1])
		value.SetMapIndex(keyValue, mapValue)
	}

	return value, nil
}

func parseSliceValue(s string, refValue reflect.Value) (reflect.Value, error) {
	value := refValue
	sSlice := strings.Split(s, ";")
	if len(sSlice) == 0 {
		return value, errors.New("empty value")
	}
	for _, v := range sSlice {
		sliceValue := reflect.Indirect(reflect.New(refValue.Addr().Type().Elem().Elem()))
		err := setValue(sliceValue, v)
		if err != nil {
			return value, err
		}
		value = reflect.Append(value, sliceValue)
	}

	return value, nil
}

func validConfKeys(t reflect.Type, to map[string]string, pk, pv string) {
	for i := 0; i < t.NumField(); i++ {
		name := strings.TrimSpace(strings.Split(t.Field(i).Tag.Get("json"), ",")[0])
		if name == "" {
			name = t.Field(i).Name
		}
		if pk != "" {
			name = pk + "." + name
		}
		kt := t.Field(i).Name
		if pv != "" {
			kt = pv + "." + kt
		}
		if t.Field(i).Type.Kind() == reflect.Struct {
			validConfKeys(t.Field(i).Type, to, name, kt)
		} else if t.Field(i).Type.Kind() == reflect.Slice {
			validConfKeys(t.Field(i).Type.Elem(), to, name, kt)
		} else {
			to[strings.ToLower(name)] = kt
		}
	}
}