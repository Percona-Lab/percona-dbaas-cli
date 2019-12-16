package options

import (
	"log"
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

// TODO: maps, slices
func getValue(val reflect.Value, value string) (reflect.Value, error) {
	if val.Kind() == reflect.Ptr {
		iv, err := getValue(val.Elem(), value)
		if err != nil {
			return reflect.Value{}, err
		}
		reflect.Indirect(val).Set(iv)
		return val, nil
	}

	rv := reflect.Value{}
	switch val.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil || val.OverflowInt(v) {
			return rv, errors.Errorf("parse value %s: %v", val, err)
		}
		rv = reflect.ValueOf(v).Convert(val.Type())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		v, err := strconv.ParseUint(value, 10, 64)
		if err != nil || val.OverflowUint(v) {
			return rv, errors.Errorf("parse value %s: %v", val, err)
		}
		rv = reflect.ValueOf(v).Convert(val.Type())
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(value, 64)
		if err != nil || val.OverflowFloat(v) {
			return rv, errors.Errorf("parse value %s: %v", val, err)
		}
		rv = reflect.ValueOf(v).Convert(val.Type())
	case reflect.Bool:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return rv, errors.Errorf("parse value %s: %v", val, err)
		}
		rv = reflect.ValueOf(v)
	case reflect.String:
		rv = reflect.ValueOf(value)
	case reflect.Map:
		//v, err := parseMapValue(value)
		//if err != nil {
		//	return rv, errors.Errorf("parse value %s: %v", val, err)
		//}
		//rv = reflect.ValueOf(v)
	}

	return rv, nil
}

func parseMapValue(s string, refValue reflect.Value) (reflect.Value, error) {
	value := refValue
	value.Set(reflect.MakeMap(refValue.Type()))
	//value := make(map[interface{}]interface{})
	sSlice := strings.Split(s, ";")
	if len(sSlice) == 0 {
		return value, errors.New("empty value")
	}
	for _, v := range sSlice {
		vSlice := strings.Split(v, ":")
		if len(vSlice) != 2 {
			return value, errors.New("empty map value")
		}
		key, err := getValue(refValue, vSlice[0])
		if err != nil {
			return value, errors.Wrap(err, "get map key")
		}
		val, err := getValue(refValue, vSlice[1])
		if err != nil {
			return value, errors.Wrap(err, "get map value")
		}
		log.Println(value.Type().String())
		value.SetMapIndex(key, val)
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
		} else {
			to[name] = kt
		}
	}
}
