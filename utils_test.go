package textmagic

import (
	"reflect"
	"testing"
)

func TestSplitSlice(t *testing.T) {
	defer func(oldMaxInSlice int) {
		maxInSlice = oldMaxInSlice
	}(maxInSlice)
	maxInSlice = 3
	tests := []struct {
		input  []string
		output [][]string
	}{
		{[]string{}, [][]string{}},
		{[]string{"1"}, [][]string{{"1"}}},
		{[]string{"1", "2"}, [][]string{{"1", "2"}}},
		{[]string{"1", "2", "3"}, [][]string{{"1", "2", "3"}}},
		{[]string{"1", "2", "3", "4"}, [][]string{{"1", "2", "3"}, {"4"}}},
		{[]string{"1", "2", "3", "4", "5"}, [][]string{{"1", "2", "3"}, {"4", "5"}}},
		{[]string{"1", "2", "3", "4", "5", "6"}, [][]string{{"1", "2", "3"}, {"4", "5", "6"}}},
		{[]string{"1", "2", "3", "4", "5", "6", "7"}, [][]string{{"1", "2", "3"}, {"4", "5", "6"}, {"7"}}},
	}
	for n, test := range tests {
		if got := splitSlice(test.input); !reflect.DeepEqual(got, test.output) {
			t.Errorf("test %d: expecting %v, got %v", n+1, test.output, got)
		}
	}
}

func TestUtos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input  uint64
		output string
	}{
		{0, "0"},
		{1, "1"},
		{2, "2"},
		{9, "9"},
		{10, "10"},
		{11, "11"},
		{15, "15"},
		{16, "16"},
		{17, "17"},
		{20, "20"},
		{99, "99"},
		{100, "100"},
		{999, "999"},
		{1000, "1000"},
		{9999, "9999"},
		{10000, "10000"},
		{18446744073709551615, "18446744073709551615"},
	}
	for n, test := range tests {
		if got := utos(test.input); got != test.output {
			t.Errorf("test %d: expecting %s, got %s", n+1, test.output, got)
		}
	}
}
