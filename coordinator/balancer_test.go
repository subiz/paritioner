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
		expected map[string][]int32
	}{{
		desc: "test 1, remove node",
		config: map[string][]int32{
			"1": []int32{1, 2},
			"2": []int32{3, 4},
			"3": []int32{5, 6},
			"4": []int32{7, 8},
		},
		nodes: []string{"1", "2", "3"},
		expected: map[string][]int32{
			"1": []int32{1, 2, 7},
			"2": []int32{3, 4, 8},
			"3": []int32{5, 6},
		},
	}, {
		desc: "test 2, add node",
		config: map[string][]int32{
			"1": []int32{1, 2},
			"2": []int32{3, 4},
			"3": []int32{5, 6},
			"4": []int32{7, 8},
		},
		nodes: []string{"1", "2", "3", "4", "5"},
		expected: map[string][]int32{
			"1": []int32{1, 2},
			"2": []int32{3, 4},
			"3": []int32{5, 6},
			"4": []int32{7},
			"5": []int32{8},
		},
	}, {
		desc: "test 4, no change in cluster, just rebalance",
		config: map[string][]int32{
			"1": []int32{1, 2},
			"2": []int32{3, 4},
			"3": []int32{5, 6},
			"4": []int32{},
		},
		nodes: []string{"1", "2", "3", "4"},
		expected: map[string][]int32{
			"1": []int32{1, 2},
			"2": []int32{3, 4},
			"3": []int32{5},
			"4": []int32{6},
		},
	}, {
		desc: "unbalanced remove node",
		config: map[string][]int32{
			"1": []int32{1, 2, 3, 4, 5, 9},
			"2": []int32{6},
			"3": []int32{7, 8},
			"4": []int32{},
		},
		nodes: []string{"1", "3", "4"},
		expected: map[string][]int32{
			"1": []int32{1, 2, 3},
			"3": []int32{7, 8, 4},
			"4": []int32{5, 9, 6},
		},
	}, {
		desc: "unbalanced add node",
		config: map[string][]int32{
			"1": []int32{1, 2, 3, 4, 5, 6, 9},
			"3": []int32{7, 8},
		},
		nodes: []string{"1", "2", "3"},
		expected: map[string][]int32{
			"1": []int32{1, 2, 3},
			"2": []int32{9, 5, 6},
			"3": []int32{7, 8, 4},
		},
	}, {
		desc: "remove many nodes",
		config: map[string][]int32{
			"1": []int32{1, 2, 3, 4, 5},
			"2": []int32{6, 7, 8, 9, 10},
			"3": []int32{11, 12, 13, 14, 15},
			"4": []int32{16, 17, 18, 19, 20},
		},
		nodes: []string{"1", "2"},
		expected: map[string][]int32{
			"1": []int32{1, 2, 3, 4, 5, 11, 12, 13, 14, 15},
			"2": []int32{6, 7, 8, 9, 10, 16, 17, 18, 19, 20},
		},
	}}

	for _, tc := range tcs {
		out := balance(tc.config, tc.nodes)
		if !compareM(out, tc.expected) {
			t.Fatalf("in test '%s', expect %v, got %v", tc.desc, tc.expected, out)
		}
	}
}
