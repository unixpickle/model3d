package toolbox3d

import (
	"flag"
	"testing"
)

type TestAddFlagsSubObj struct {
}

type TestAddFlagsObj struct {
	IntField      int
	OtherIntField int     `default:"123"`
	FloatField    float64 `default:"3.14"`
}

func TestAddFlags(t *testing.T) {
	var obj TestAddFlagsObj
	fs := flag.NewFlagSet("foo", flag.PanicOnError)
	AddFlags(&obj, fs)
	fs.Parse([]string{})
	if obj.OtherIntField != 123 {
		t.Errorf("incorrect OtherIntField: %v", obj.OtherIntField)
	}
	if obj.IntField != 0 {
		t.Errorf("incorrect IntField: %v", obj.IntField)
	}
	if obj.FloatField != 3.14 {
		t.Errorf("incorrect FloatField: %v", obj.FloatField)
	}

	var obj1 TestAddFlagsObj
	fs = flag.NewFlagSet("foo", flag.PanicOnError)
	AddFlags(&obj1, fs)
	fs.Parse([]string{"-int-field", "4", "-other-int-field", "5", "-float-field", "3.14"})
	if obj1.OtherIntField != 5 {
		t.Errorf("incorrect OtherIntField: %v", obj.OtherIntField)
	}
	if obj1.IntField != 4 {
		t.Errorf("incorrect IntField: %v", obj.IntField)
	}
	if obj1.FloatField != 3.14 {
		t.Errorf("incorrect FloatField: %v", obj.FloatField)
	}
}
