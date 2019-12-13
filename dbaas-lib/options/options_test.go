package options_test

import (
	"reflect"
	"testing"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

func TestOptions(t *testing.T) {
	type T2 struct {
		StringVal  string  `json:"s1"`
		StringValP *string `json:"s2"`
		IntVal     int64   `json:"i1"`
		IntValP    *int32  `json:"i2"`
	}
	type T struct {
		Level0 T2 `json:"l"`
	}

	i32 := int32(666)
	str := "value2"
	cmp := T{
		Level0: T2{
			StringVal:  "value1",
			StringValP: &str,
			IntVal:     42,
			IntValP:    &i32,
		},
	}

	v := T{}

	err := options.Parse(&v, reflect.TypeOf(v), "l.s1=value1,l.s2=value2,l.i1=42,l.i2=666")
	if err != nil {
		t.Errorf("Parse error: %v", err)
	}

	if !reflect.DeepEqual(cmp, v) {
		t.Errorf("not equal: %v", v)
	}
}
