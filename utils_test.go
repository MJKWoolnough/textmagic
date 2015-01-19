package textmagic

import (
	"reflect"
	"testing"
)

func TestJoinUints(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input  []uint64
		output string
	}{
		{[]uint64{0}, "0"},
		{[]uint64{1}, "1"},
		{[]uint64{1, 2}, "1,2"},
		{[]uint64{1, 2, 3, 4, 5}, "1,2,3,4,5"},
		{[]uint64{9876543210, 1234567890, 0}, "9876543210,1234567890,0"},
	}
	for n, test := range tests {
		if got := joinUints(test.input); got != test.output {
			t.Errorf("test %d: expecting %s, got %s", n+1, test.output, got)
		}
	}
}

func TestSplitSlice(t *testing.T) {
	defer func(oldMaxInSlice int) {
		maxInSlice = oldMaxInSlice
	}(maxInSlice)
	maxInSlice = 3
	tests := []struct {
		input  []uint64
		output [][]uint64
	}{
		{[]uint64{}, [][]uint64{}},
		{[]uint64{1}, [][]uint64{{1}}},
		{[]uint64{1, 2}, [][]uint64{{1, 2}}},
		{[]uint64{1, 2, 3}, [][]uint64{{1, 2, 3}}},
		{[]uint64{1, 2, 3, 4}, [][]uint64{{1, 2, 3}, {4}}},
		{[]uint64{1, 2, 3, 4, 5}, [][]uint64{{1, 2, 3}, {4, 5}}},
		{[]uint64{1, 2, 3, 4, 5, 6}, [][]uint64{{1, 2, 3}, {4, 5, 6}}},
		{[]uint64{1, 2, 3, 4, 5, 6, 7}, [][]uint64{{1, 2, 3}, {4, 5, 6}, {7}}},
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

func TestStou(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input   string
		outputb bool
		outputu uint64
	}{
		{"0", true, 0},
		{"1", true, 1},
		{"9", true, 9},
		{"10", true, 10},
		{"15", true, 15},
		{"16", true, 16},
		{"17", true, 17},
		{"20", true, 20},
		{"99", true, 99},
		{"100", true, 100},
		{"18446744073709551615", true, 18446744073709551615},
		{"a", false, 0},
		{"0a", false, 0},
		{"243423a", false, 0},
	}
	for n, test := range tests {
		if gotu, gotb := stou(test.input); test.outputb && !gotb {
			t.Errorf("test %d: expecting a valid uint, got false", n+1)
		} else if !test.outputb && gotb {
			t.Errorf("test %d: expecting a non-valid uint, got true", n+1)
		} else if test.outputb && gotb && gotu != test.outputu {
			t.Errorf("test %d: expecting value %d, got %d", n+1, test.outputu, gotu)
		}
	}
}
