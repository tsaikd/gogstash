package kv

import (
	"reflect"
	"testing"
)

func Test_splitQuotedStringsBySpace(t *testing.T) {
	// test with three values
	input := "val1=number1 val2=\"hello world\" val3=123"
	result := splitQuotedStringsBySpace(input)
	if len(result) != 3 {
		t.Error("Should have received three elements, got", result, len(result))
	}
	// test without any values
	result = splitQuotedStringsBySpace("")
	if len(result) != 0 {
		t.Error("Should have received 0 elements, got", result, len(result))
	}
	// test with invalid input
	result = splitQuotedStringsBySpace("k=v hi")
	if len(result) != 1 || result[0] != "k=v" {
		t.Error("Should have received 1 valid element, got", result, len(result))
	}
	// test with invalid input in beginning
	result = splitQuotedStringsBySpace("invalid k=v")
	if len(result) != 1 || result[0] != "k=v" {
		t.Error("Should have received 1 valid element, got", result, len(result))
	}
	// test with one leading space
	result = splitQuotedStringsBySpace(" k=v")
	if len(result) != 1 || result[0] != "k=v" {
		t.Errorf("Failed on one leading space, got '%s'", result[0])
	}
	// test with one trailing space
	result = splitQuotedStringsBySpace("k=v ")
	if len(result) != 1 || result[0] != "k=v" {
		t.Errorf("Failed on one trailing space, got '%s'", result[0])
	}
	// test two elements two double spaces between them
	result = splitQuotedStringsBySpace("xk=vx  xx=yyy")
	if len(result) != 2 || result[0] != "xk=vx" {
		t.Error("Failed on double spaces", result, len(result))
	}
}

func Test_splitIntoKV(t *testing.T) {
	// check
	input := []string{"key=value", "number=1622940831444131311", "float=2.54", "smallnumber=3"}
	expNum := 1622940831444131311
	result := splitIntoKV(input, []string{"smallnumber"})
	// check that we have the right number of elements
	if len(result) != len(input) {
		t.Errorf("Expected %d elements, got %d", len(input), len(result))
	}
	// check that our number is of type int
	switch result["number"].(type) {
	case uint, uint64, int, int64:
		ourNum := result["number"].(int)
		if expNum != ourNum {
			t.Errorf("Number should be %d, got %d", expNum, ourNum)
		}
	default:
		t.Error("Number was not of integer, was of type", reflect.TypeOf(result["number"]))
	}
	// check that smallnumber is a string
	if reflect.TypeOf(result["smallnumber"]) != reflect.TypeOf("") {
		t.Error("smallnumber not string, got", reflect.TypeOf(result["smallnumber"]))
	}
	// check that float is of string
	typeof := reflect.TypeOf(result["float"])
	if typeof != reflect.TypeOf("string") {
		t.Error("Float was not of type string, got", typeof)
	}
	// check invalid input
	input = []string{"ID=", "val", "A=B", "AA= "}
	result = splitIntoKV(input, []string{})
	if len(result) != 2 {
		t.Error("splitIntoKV failed with invalid input")
	}
}

func Test_contains(t *testing.T) {
	type args struct {
		key      string
		elements *[]string
	}
	myList := &[]string{"A", "AAA", "BB"}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Check A", args{"A", myList}, true},
		{"Check B", args{"B", myList}, false},
		{"Check empty list", args{"B", &[]string{}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.args.key, tt.args.elements); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
