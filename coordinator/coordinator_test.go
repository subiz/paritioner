package coordinator

import (
	"sort"
	"testing"
)

func compareIntArr(a1, a2 []int) bool {
	if len(a1) != len(a2) {
		return false
	}

	sort.Ints(a1)
	sort.Ints(a2)

	for i := range a1 {
		if a1[i] != a2[i] {
			return false
		}
	}
	return true
}

func compareM(c1, c2 map[string][]int) bool {
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
		config   map[string][]int
		nodes    []string
		expected map[string][]int
	}{{
		desc: "test 1, remove node",
		config: map[string][]int{
			"1": []int{1, 2},
			"2": []int{3, 4},
			"3": []int{5, 6},
			"4": []int{7, 8},
		},
		nodes: []string{"1", "2", "3"},
		expected: map[string][]int{
			"1": []int{1, 2, 7},
			"2": []int{3, 4, 8},
			"3": []int{5, 6},
		},
	}, {
		desc: "test 2, add node",
		config: map[string][]int{
			"1": []int{1, 2},
			"2": []int{3, 4},
			"3": []int{5, 6},
			"4": []int{7, 8},
		},
		nodes: []string{"1", "2", "3", "4", "5"},
		expected: map[string][]int{
			"1": []int{1, 2},
			"2": []int{3, 4},
			"3": []int{5, 6},
			"4": []int{7},
			"5": []int{8},
		},
	}, {
		desc: "test 4, no change in cluster, just rebalance",
		config: map[string][]int{
			"1": []int{1, 2},
			"2": []int{3, 4},
			"3": []int{5, 6},
			"4": []int{},
		},
		nodes: []string{"1", "2", "3", "4"},
		expected: map[string][]int{
			"1": []int{1, 2},
			"2": []int{3, 4},
			"3": []int{5},
			"4": []int{6},
		},
	}, {
		desc: "unbalanced remove node",
		config: map[string][]int{
			"1": []int{1, 2, 3, 4, 5, 9},
			"2": []int{6},
			"3": []int{7, 8},
			"4": []int{},
		},
		nodes: []string{"1", "3", "4"},
		expected: map[string][]int{
			"1": []int{1, 2, 3},
			"3": []int{7, 8, 4},
			"4": []int{5, 9, 6},
		},
	}, {
		desc: "unbalanced add node",
		config: map[string][]int{
			"1": []int{1, 2, 3, 4, 5, 6, 9},
			"3": []int{7, 8},
		},
		nodes: []string{"1", "2", "3"},
		expected: map[string][]int{
			"1": []int{1, 2, 3},
			"2": []int{9, 5, 6},
			"3": []int{7, 8, 4},
		},
	}}

	for _, tc := range tcs {
		out := balance(tc.config, tc.nodes)
		if !compareM(out, tc.expected) {
			t.Fatalf("in test '%s', expect %v, got %v", tc.desc, tc.expected, out)
		}
	}
}
