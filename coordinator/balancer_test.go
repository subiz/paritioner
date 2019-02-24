package main

import (
	"testing"
)

func compareIntArr(a1, a2 []int32) bool {
	if len(a1) != len(a2) {
		return false
	}

	m := make(map[int32]bool, len(a1))

	for _, i := range a1 {
		m[i] = true
	}

	for _, i := range a2 {
		if _, ok := m[i]; !ok {
			return false
		}
	}
	return true
}

func compareM(c1, c2 map[string][]int32) bool {
	if len(c1) != len(c2) {
		return false
	}

	for k, v1 := range c1 {
		v2, ok := c2[k]
		if !ok {
			return false
		}

		if !compareIntArr(v1, v2) {
			return false
		}
	}
	return true
}

func TestBalance(t *testing.T) {
	tcs := []struct {
		desc     string
		config   map[string][]int32
		nodes    []string
		numPars  int32
		expected map[string][]int32
	}{{
		desc: "test 1, remove node",
		config: map[string][]int32{
			"1": []int32{0, 1},
			"2": []int32{2, 3},
			"3": []int32{4, 5},
			"4": []int32{6, 7},
		},
		numPars: 8,
		nodes:   []string{"1", "2", "3"},
		expected: map[string][]int32{
			"1": []int32{0, 1, 6},
			"2": []int32{2, 3, 7},
			"3": []int32{4, 5},
		},
	}, {
		desc: "test 2, add node",
		config: map[string][]int32{
			"1": []int32{0, 1},
			"2": []int32{2, 3},
			"3": []int32{4, 5},
			"4": []int32{6, 7},
		},
		numPars: 8,
		nodes:   []string{"1", "2", "3", "4", "5"},
		expected: map[string][]int32{
			"1": []int32{0, 1},
			"2": []int32{2, 3},
			"3": []int32{4, 5},
			"4": []int32{6},
			"5": []int32{7},
		},
	}, {
		desc: "test 4, no change in cluster, just rebalance",
		config: map[string][]int32{
			"1": []int32{0, 1},
			"2": []int32{2, 3},
			"3": []int32{4, 5},
			"4": []int32{},
		},
		numPars: 6,
		nodes:   []string{"1", "2", "3", "4"},
		expected: map[string][]int32{
			"1": []int32{0, 1},
			"2": []int32{2, 3},
			"3": []int32{4},
			"4": []int32{5},
		},
	}, {
		desc: "unbalanced remove node",
		config: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4, 8},
			"2": []int32{5},
			"3": []int32{6, 7},
			"4": []int32{},
		},
		numPars: 9,
		nodes:   []string{"1", "3", "4"},
		expected: map[string][]int32{
			"1": []int32{0, 1, 2},
			"3": []int32{6, 7, 3},
			"4": []int32{4, 8, 5},
		},
	}, {
		desc: "unbalanced add node",
		config: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4, 5, 8},
			"3": []int32{6, 7},
		},
		numPars: 9,
		nodes:   []string{"1", "2", "3"},
		expected: map[string][]int32{
			"1": []int32{0, 1, 2},
			"2": []int32{8, 4, 5},
			"3": []int32{6, 7, 3},
		},
	}, {
		desc: "remove many nodes",
		config: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4},
			"2": []int32{5, 6, 7, 8, 9},
			"3": []int32{10, 11, 12, 13, 14},
			"4": []int32{15, 16, 17, 18, 19},
		},
		numPars: 20,
		nodes:   []string{"1", "2"},
		expected: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4, 10, 11, 12, 13, 14},
			"2": []int32{5, 6, 7, 8, 9, 15, 16, 17, 18, 19},
		},
	}, {
		desc: "remove all nodes",
		config: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4},
			"2": []int32{5, 6, 7, 8, 9},
			"3": []int32{10, 11, 12, 13, 14},
			"4": []int32{15, 16, 17, 18, 19},
		},
		numPars:  20,
		nodes:    []string{},
		expected: map[string][]int32{},
	}, {
		desc:    "join all nodes",
		config:  map[string][]int32{},
		numPars: 20,
		nodes:   []string{"1", "2", "3", "4"},
		expected: map[string][]int32{
			"1": []int32{0, 1, 2, 3, 4},
			"2": []int32{5, 6, 7, 8, 9},
			"3": []int32{10, 11, 12, 13, 14},
			"4": []int32{15, 16, 17, 18, 19},
		},
	}}

	for _, tc := range tcs {
		out := balance(tc.numPars, tc.config, tc.nodes)
		if !compareM(out, tc.expected) {
			t.Fatalf("in test '%s', expect %v, got %v", tc.desc, tc.expected, out)
		}
	}
}
