package pxc

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/engines/k8s-pxc/types/config"
	"github.com/pkg/errors"
)

var currectOptions map[string]string

// ParseOptions parse PXC options given in "object.paramValue=val,objectTwo.paramValue=val" string
func (p *PXC) ParseOptions(options string) error {
	var c config.ClusterConfig

	res := config.PodResources{
		Requests: config.ResourcesList{
			CPU:    "600m",
			Memory: "1G",
		},
	}
	topologyKey := "none" //TODO: Deside what value is default "none" or "kubernetes.io/hostname"
	aff := config.PodAffinity{
		TopologyKey: topologyKey,
	}
	c.PXC.Size = int32(3)
	c.PXC.Resources = res
	c.PXC.Affinity = aff
	c.ProxySQL.Size = int32(1)
	c.ProxySQL.Resources = res
	c.ProxySQL.Affinity = aff
	c.S3.SkipStorage = true

	if len(options) != 0 {
		currectOptions = make(map[string]string)
		keys(reflect.TypeOf(config.ClusterConfig{}), "", "")

		optArr := strings.Split(options, ",")

		for _, str := range optArr {
			v := strings.Split(str, "=")
			if _, ok := currectOptions[v[0]]; !ok {
				return errors.New("incorrect options")
			}
			if len(v) > 1 {
				err := set(&c, currectOptions[v[0]], v[1])
				if err != nil {
					return errors.Wrap(err, "set value")
				}
			}
		}
	}
	p.config = c

	return nil
}

func set(i *config.ClusterConfig, field string, value string) error {
	fs := strings.Split(field, ".")
	v := reflect.ValueOf(i).Elem()

	switch len(fs) {
	case 1:
		val, err := getValue(value, v.FieldByName(fs[0]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).Set(val)
	case 2:
		val, err := getValue(value, v.FieldByName(fs[0]).FieldByName(fs[1]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).Set(val)
	case 3:
		val, err := getValue(value, v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).Set(val)
	case 4:
		val, err := getValue(value, v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]).Set(val)
	case 5:
		val, err := getValue(value, v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]).FieldByName(fs[4]))
		if err != nil {
			return err
		}
		v.FieldByName(fs[0]).FieldByName(fs[1]).FieldByName(fs[2]).FieldByName(fs[3]).FieldByName(fs[4]).Set(val)
	}
	return nil
}

func getValue(value string, field reflect.Value) (reflect.Value, error) {
	var fieldValue interface{}
	var pointer bool
	if strings.Contains(field.Type().String(), "*") {
		pointer = true
	}
	fieldType := strings.Trim(field.Type().String(), "*")
	switch fieldType {
	case "int":
		v, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		fieldValue = v
		if pointer {
			fieldValue = &v
		}
	case "int32":
		v, err := strconv.Atoi(value)
		if err != nil {
			return reflect.Value{}, err
		}
		i32v := int32(v)
		fieldValue = i32v
		if pointer {
			fieldValue = &i32v
		}
	case "int64":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		fieldValue = v
		if pointer {
			fieldValue = &v
		}
	case "float32":
		v, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		f32v := float32(v)
		fieldValue = f32v
		if pointer {
			fieldValue = &f32v
		}
	case "float64":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		fieldValue = v
		if pointer {
			fieldValue = &v
		}
	case "bool":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return reflect.Value{}, err
		}
		fieldValue = v
		if pointer {
			fieldValue = &v
		}
	case "map[string]string":
		v, err := parseMapValue(value)
		if err != nil {
			return reflect.Value{}, err
		}
		fieldValue = v
		if pointer {
			fieldValue = &v
		}
		return reflect.Indirect(reflect.ValueOf(v)), nil
	case "string":
		fieldValue = value
		if pointer {
			fieldValue = &value
		}
	}

	return reflect.ValueOf(fieldValue), nil
}

func parseMapValue(s string) (map[string]string, error) {
	value := make(map[string]string)
	sSlice := strings.Split(s, ";")
	if len(sSlice) == 0 {
		return nil, errors.New("empty value")
	}
	for _, v := range sSlice {
		vSlice := strings.Split(v, ":")
		if len(vSlice) != 2 {
			return nil, errors.New("empty map value")
		}
		value[vSlice[0]] = vSlice[1]
	}

	return value, nil
}

func keys(t reflect.Type, prevType, prevName string) map[string]string {
	var v = make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		name := strings.TrimSpace(strings.Split(t.Field(i).Tag.Get("json"), ",")[0])
		if t.Field(i).Type.Kind() == reflect.Struct {
			for name, nType := range keys(t.Field(i).Type, prevType+t.Field(i).Name+".", prevName+name+".") {
				currectOptions[prevName+name+"."+name] = prevType + t.Field(i).Name + "." + nType
			}
		} else {
			v[name] = t.Field(i).Name
			currectOptions[prevName+name] = prevType + t.Field(i).Name
		}
	}

	return v
}
