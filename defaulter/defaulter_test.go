package defaulter

import (
	"testing"
)

var testData = map[string]any{
	"float": 15.6,
}

type TestStruct struct {
	StringField   string  `default:"hello"`
	IntField      int     `default:"42"`
	BoolField     bool    `default:"true"`
	FloatField    float64 `default:"3.14"`
	FloatTemplate float64 `default:"{{ .float }}"`
	NoDefault     int
	StructField   NestedStruct
	PointerField  *NestedStruct
	Fields        []NestedStruct
	PointerFieldS []*NestedStruct
}

type NestedStruct struct {
	Value string `default:"\"nested\""`
}

func TestEvaluateOnStruct(t *testing.T) {
	s := &TestStruct{
		PointerField: &NestedStruct{},
	}

	ApplyDefaults(s, testData)

	if s.StringField != "hello" {
		t.Errorf("StringField: expected \"hello\", got %q", s.StringField)
	}
	if s.IntField != 42 {
		t.Errorf("IntField: expected 42, got %d", s.IntField)
	}
	if s.BoolField != true {
		t.Errorf("BoolField: expected true, got %v", s.BoolField)
	}
	if s.FloatField != 3.14 {
		t.Errorf("FloatField: expected 3.14, got %f", s.FloatField)
	}
	if s.FloatTemplate != 15.6 {
		t.Errorf("FloatTemplate: expected 15.6, got %f", s.FloatTemplate)
	}
	if s.StructField.Value != "nested" {
		t.Errorf("StructField.Value: expected \"nested\", got %q", s.StructField.Value)
	}
	if s.PointerField.Value != "nested" {
		t.Errorf("PointerField.Value: expected \"nested\", got %q", s.PointerField.Value)
	}
}

func TestEvaluateOnStructWithExistingValues(t *testing.T) {
	s := &TestStruct{
		StringField: "existing",
		IntField:    100,
	}

	ApplyDefaults(s, nil)

	if s.StringField != "existing" {
		t.Errorf("StringField should not be overwritten: expected \"existing\", got %q", s.StringField)
	}
	if s.IntField != 100 {
		t.Errorf("IntField should not be overwritten: expected 100, got %d", s.IntField)
	}
}
