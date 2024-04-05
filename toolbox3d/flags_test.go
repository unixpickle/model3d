package toolbox3d

import (
	"flag"
	"testing"
	"time"
)

type TestAddFlagsSubObj struct {
}

type TestAddFlagsObj struct {
	IntField      int
	OtherIntField int           `default:"123"`
	FloatField    float64       `default:"3.14"`
	Int64Field    int64         `default:"123123"`
	Delay         time.Duration `default:"3s"`
	UintField     uint          `default:"333"`
	Uint64Field   uint64        `default:"444"`
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
	if obj.Int64Field != 123123 {
		t.Errorf("incorrect Int64Field: %v", obj.Int64Field)
	}
	if obj.Delay != time.Second*3 {
		t.Errorf("incorrect Delay: %v", obj.Delay)
	}
	if obj.UintField != 333 {
		t.Errorf("incorrect UintField: %v", obj.UintField)
	}
	if obj.Uint64Field != 444 {
		t.Errorf("incorrect Uint64Field: %v", obj.Uint64Field)
	}

	var obj1 TestAddFlagsObj
	fs = flag.NewFlagSet("foo", flag.PanicOnError)
	AddFlags(&obj1, fs)
	fs.Parse([]string{"-int-field", "4", "-other-int-field", "5", "-float-field", "3.14",
		"-int64-field=-64", "-delay", "5m", "-uint-field", "331", "-uint64-field", "441"})
	if obj1.OtherIntField != 5 {
		t.Errorf("incorrect OtherIntField: %v", obj1.OtherIntField)
	}
	if obj1.IntField != 4 {
		t.Errorf("incorrect IntField: %v", obj1.IntField)
	}
	if obj1.FloatField != 3.14 {
		t.Errorf("incorrect FloatField: %v", obj1.FloatField)
	}
	if obj1.Int64Field != -64 {
		t.Errorf("incorrect Int64Field: %v", obj1.Int64Field)
	}
	if obj1.Delay != time.Minute*5 {
		t.Errorf("incorrect Delay: %v", obj1.Delay)
	}
	if obj1.UintField != 331 {
		t.Errorf("incorrect UintField: %v", obj1.UintField)
	}
	if obj1.Uint64Field != 441 {
		t.Errorf("incorrect Uint64Field: %v", obj1.Uint64Field)
	}
}
