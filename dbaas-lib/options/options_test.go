package options_test

import (
	"reflect"
	"testing"

	"github.com/Percona-Lab/percona-dbaas-cli/dbaas-lib/options"
)

type customInt int64

func TestOptions(t *testing.T) {
	type T3 struct {
		StringValP  *string `json:"sp1"`
		StringValP2 *string `json:"sp2"`
	}
	type T2 struct {
		StringVal  string         `json:"s1"`
		StringValP *string        `json:"s2"`
		IntVal     int64          `json:"i1"`
		IntValP    *int32         `json:"i2"`
		MapVal     map[string]int `json:"m1"`
		SliceVal   []int32        `json:"sl1"`
		CustomType customInt      `json:"c1"`
	}
	type T struct {
		Level0 T2  `json:"l"`
		Level1 *T3 `json:"l1"`
	}

	i32 := int32(666)
	str := "value2"
	mapVal := map[string]int{"test": 1}
	var customVar customInt
	customVar = 2356
	t3 := T3{
		StringValP:  &str,
		StringValP2: &str,
	}
	cmp := T{
		Level0: T2{
			StringVal:  "value1",
			StringValP: &str,
			IntVal:     42,
			IntValP:    &i32,
			MapVal:     mapVal,
			SliceVal:   []int32{1, 2},
			CustomType: customVar,
		},
		Level1: &t3,
	}

	v := T{}

	err := options.Parse(&v, reflect.TypeOf(v), "L.S1=value1,l.s2=value2,l.i1=42,l.i2=666,l.m1=test:1,l.sl1=1;2,l.c1=2356,l1.sp1=value2,l1.sp2=value2")
	if err != nil {
		t.Errorf("Parse error: %v", err)
	}

	if !reflect.DeepEqual(cmp, v) {
		t.Errorf("not equal: %v", v)
	}
}
